// Shared utilities for vscode-textmate-based fixture generation.
// Modeled after giallo's scripts/lib/textmate-common.js.

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const require = createRequire(import.meta.url);

const vsctmRoot = path.resolve(__dirname, '..', '..', '..', 'vscode-textmate');
const { Registry, parseRawGrammar } = require(path.join(vsctmRoot, 'out', 'src', 'main'));
const { EncodedTokenAttributes } = require(path.join(vsctmRoot, 'out', 'src', 'encodedTokenAttributes'));

const grammarsThemesRoot = path.resolve(__dirname, '..', '..', '..', 'grammars-themes');
const GRAMMAR_DIR = path.join(grammarsThemesRoot, 'packages', 'tm-grammars', 'grammars');
const THEMES_DIR = path.join(grammarsThemesRoot, 'packages', 'tm-themes', 'themes');

async function getOniguruma() {
  const vscodeOniguruma = require(path.join(vsctmRoot, 'node_modules', 'vscode-oniguruma'));
  const wasmBin = fs.readFileSync(
    path.join(vsctmRoot, 'node_modules', 'vscode-oniguruma', 'release', 'onig.wasm')
  ).buffer;
  await vscodeOniguruma.loadWASM(wasmBin);
  return {
    createOnigScanner(patterns) { return new vscodeOniguruma.OnigScanner(patterns); },
    createOnigString(s) { return new vscodeOniguruma.OnigString(s); },
  };
}

export async function createRegistry() {
  const loadedGrammars = new Map();
  const nameToScope = new Map();

  const files = fs.readdirSync(GRAMMAR_DIR);
  for (const file of files) {
    const ext = path.extname(file).toLowerCase();
    if (ext !== '.json') continue;
    const baseName = path.basename(file, ext).toLowerCase();
    const fullPath = path.join(GRAMMAR_DIR, file);
    try {
      const content = fs.readFileSync(fullPath, 'utf-8');
      const rawGrammar = parseRawGrammar(content, fullPath);
      if (rawGrammar && rawGrammar.scopeName) {
        loadedGrammars.set(rawGrammar.scopeName, rawGrammar);
        nameToScope.set(baseName, rawGrammar.scopeName);
      }
    } catch (err) {
      console.warn(`  WARN: failed to parse grammar ${baseName}: ${err.message}`);
    }
  }

  const onigLib = await getOniguruma();
  const registry = new Registry({
    onigLib,
    loadGrammar: async (scopeName) => loadedGrammars.get(scopeName) || null,
  });

  for (const [scopeName] of loadedGrammars) {
    try {
      await registry.loadGrammar(scopeName);
    } catch (err) {
      console.warn(`  WARN: failed to load grammar ${scopeName}: ${err.message}`);
    }
  }

  return { registry, nameToScope, EncodedTokenAttributes };
}

export function loadTheme(themeName) {
  const themePath = path.join(THEMES_DIR, themeName + '.json');
  const raw = JSON.parse(fs.readFileSync(themePath, 'utf-8'));

  if (raw.tokenColors && !raw.settings) {
    const defaultSetting = {
      settings: {
        foreground: raw.colors && raw.colors['editor.foreground'],
      },
    };
    return {
      name: raw.name || themeName,
      settings: [defaultSetting, ...raw.tokenColors],
      colors: raw.colors,
    };
  }
  return raw;
}

export function getVsctmVersion() {
  const pkg = JSON.parse(fs.readFileSync(path.join(vsctmRoot, 'package.json'), 'utf-8'));
  return pkg.version;
}
