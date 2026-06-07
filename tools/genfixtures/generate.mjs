#!/usr/bin/env node

// Fidelity fixture generator — runs real vscode-textmate over source samples
// and writes JSON fixtures for the Go fidelity test runner.
//
// Usage: node generate.mjs [--config matrix.config.json]

import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import { createRegistry, loadTheme, getVsctmVersion } from './lib/textmate-common.mjs';

const configPath = process.argv.includes('--config')
  ? process.argv[process.argv.indexOf('--config') + 1]
  : 'matrix.config.json';

const config = JSON.parse(fs.readFileSync(configPath, 'utf-8'));
const { themes, grammars, samplesDir, outputDir } = config;

// ── UTF-16 → UTF-8 offset conversion ────────────────────────────────

/**
 * Build a mapping from UTF-16 code unit offset → UTF-8 byte offset for a line.
 * @param {string} line
 * @returns {number[]} map where map[utf16Offset] = utf8ByteOffset
 */
export function buildUtf16ToUtf8Map(line) {
  const map = [0];
  let utf8Pos = 0;
  for (let i = 0; i < line.length; ) {
    const cp = line.codePointAt(i);
    const utf16Len = cp > 0xFFFF ? 2 : 1;
    const utf8Len = cp <= 0x7F ? 1 : cp <= 0x7FF ? 2 : cp <= 0xFFFF ? 3 : 4;
    utf8Pos += utf8Len;
    for (let j = 0; j < utf16Len; j++) {
      map.push(utf8Pos);
    }
    i += utf16Len;
  }
  return map;
}

function utf16ToUtf8(map, utf16Offset) {
  if (utf16Offset >= map.length) return map[map.length - 1];
  return map[utf16Offset];
}

// ── Source file loading ──────────────────────────────────────────────

function findSampleFile(grammar) {
  for (const ext of ['.sample', '.txt']) {
    const p = path.join(samplesDir, grammar + ext);
    if (fs.existsSync(p)) return p;
  }
  const subdir = path.join(samplesDir, grammar);
  if (fs.existsSync(subdir) && fs.statSync(subdir).isDirectory()) {
    const files = fs.readdirSync(subdir);
    if (files.length > 0) return path.join(subdir, files[0]);
  }
  return null;
}

// ── Font style extraction from encoded metadata ─────────────────────

function extractFontStyle(EncodedTokenAttributes, metadata) {
  const fs = EncodedTokenAttributes.getFontStyle(metadata);
  return fs || 0;
}

// ── Main ─────────────────────────────────────────────────────────────

async function main() {
  const vsctmVersion = getVsctmVersion();
  console.log(`vscode-textmate version: ${vsctmVersion}`);
  console.log(`Themes: ${themes.join(', ')}`);
  console.log(`Grammars: ${grammars.join(', ')}`);
  console.log(`Samples dir: ${path.resolve(samplesDir)}`);
  console.log(`Output dir: ${path.resolve(outputDir)}`);

  fs.mkdirSync(outputDir, { recursive: true });

  const { registry, nameToScope, EncodedTokenAttributes } = await createRegistry();

  let generated = 0;
  let skipped = 0;

  for (const grammarName of grammars) {
    const samplePath = findSampleFile(grammarName);
    if (!samplePath) {
      console.warn(`  SKIP ${grammarName}: no sample file found`);
      skipped++;
      continue;
    }

    const source = fs.readFileSync(samplePath, 'utf-8').replace(/\r\n/g, '\n');
    const lines = source.split('\n');
    if (lines.length > 0 && lines[lines.length - 1] === '') {
      lines.pop();
    }

    const grammarHash = crypto.createHash('sha256')
      .update(fs.readFileSync(samplePath))
      .digest('hex');

    const scopeName = nameToScope.get(grammarName);
    if (!scopeName) {
      console.warn(`  SKIP ${grammarName}: no scope mapping found`);
      skipped++;
      continue;
    }

    const grammar = await registry.loadGrammar(scopeName);
    if (!grammar) {
      console.warn(`  SKIP ${grammarName}: grammar failed to load`);
      skipped++;
      continue;
    }

    const fixture = {
      vsctmVersion,
      grammar: grammarName,
      grammarSourceHash: `sha256:${grammarHash}`,
      source,
      themes: {},
    };

    for (const themeName of themes) {
      const themeData = loadTheme(themeName);
      registry.setTheme(themeData);
      const colorMap = registry.getColorMap();

      let ruleStack = null;
      let binaryRuleStack = null;
      const fixtureTokens = [];

      for (let lineIdx = 0; lineIdx < lines.length; lineIdx++) {
        const lineText = lines[lineIdx];
        const lineInput = lineText + (lineIdx < lines.length - 1 ? '\n' : '');

        const t1Result = grammar.tokenizeLine(lineInput, ruleStack);
        const t2Result = grammar.tokenizeLine2(lineInput, binaryRuleStack);

        ruleStack = t1Result.ruleStack;
        binaryRuleStack = t2Result.ruleStack;

        const utf8Map = buildUtf16ToUtf8Map(lineText);
        const lineTokens = [];

        for (const t1Token of t1Result.tokens) {
          const utf16Start = t1Token.startIndex;
          const utf16End = t1Token.endIndex;
          const tokenText = lineInput.substring(utf16Start, utf16End);

          if (tokenText === '\n') continue;
          if (tokenText === '' && utf16Start === lineText.length) continue;

          const clampedEnd = Math.min(utf16End, lineText.length);
          const utf8Start = utf16ToUtf8(utf8Map, utf16Start);
          const utf8End = utf16ToUtf8(utf8Map, clampedEnd);

          let color = '';
          let fontStyle = 0;
          const matchEnd = Math.min(utf16End, lineText.length);
          for (let k = 0; k < t2Result.tokens.length / 2; k++) {
            const t2Start = t2Result.tokens[2 * k];
            const t2End = k + 1 < t2Result.tokens.length / 2
              ? t2Result.tokens[2 * (k + 1)]
              : lineInput.length;
            if (utf16Start >= t2Start && matchEnd <= t2End) {
              const metadata = t2Result.tokens[2 * k + 1];
              const fgId = EncodedTokenAttributes.getForeground(metadata);
              color = colorMap[fgId] || '';
              fontStyle = extractFontStyle(EncodedTokenAttributes, metadata);
              break;
            }
          }

          const scopes = t1Token.scopes || [];

          lineTokens.push({
            start: utf8Start,
            end: utf8End,
            text: tokenText.replace(/\n$/, ''),
            scopes,
            color,
            fontStyle,
          });
        }

        fixtureTokens.push(lineTokens);
      }

      fixture.themes[themeName] = {
        tokens: fixtureTokens,
        html: '',
      };
    }

    const sortedThemes = {};
    for (const key of Object.keys(fixture.themes).sort()) {
      sortedThemes[key] = fixture.themes[key];
    }
    fixture.themes = sortedThemes;

    const sampleBasename = path.basename(samplePath, path.extname(samplePath));
    const outName = `${grammarName}__${sampleBasename}.json`;
    const outPath = path.join(outputDir, outName);

    fs.writeFileSync(outPath, JSON.stringify(fixture, null, 2) + '\n', 'utf-8');
    console.log(`  OK ${outName} (${Object.keys(fixture.themes).length} themes)`);
    generated++;
  }

  console.log(`\nDone: ${generated} generated, ${skipped} skipped`);
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
