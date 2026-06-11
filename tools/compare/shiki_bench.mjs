import { createHighlighter } from 'shiki';
import { readFileSync } from 'node:fs';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const shikiPkg = require('shiki/package.json');
console.error(`shiki version: ${shikiPkg.version}`);

const inputsFile = process.argv[2];
const iters = parseInt(process.argv[3], 10) || 50;
const theme = process.argv[4] || 'github-dark';

const inputs = JSON.parse(readFileSync(inputsFile, 'utf-8'));
const langs = [...new Set(inputs.map(i => i.lang))];

const h = await createHighlighter({ themes: [theme], langs });

for (const input of inputs) {
  // Cold: first call (grammar already loaded by createHighlighter, but first
  // codeToHtml for this snippet includes internal first-pass costs).
  const t0 = performance.now();
  h.codeToHtml(input.code, { lang: input.lang, theme });
  const coldMs = performance.now() - t0;

  // Warm: N iterations, collect durations for median.
  const durations = [];
  for (let i = 0; i < iters; i++) {
    const t = performance.now();
    h.codeToHtml(input.code, { lang: input.lang, theme });
    durations.push(performance.now() - t);
  }
  durations.sort((a, b) => a - b);
  const warmMs = durations[Math.floor(durations.length / 2)];

  // Fidelity: token + scope counting + token dump.
  const result = h.codeToTokens(input.code, { lang: input.lang, theme, includeExplanation: true });
  let tokens = 0;
  const scopeSet = new Set();
  const dumpLines = [];

  for (const line of result.tokens) {
    tokens += line.length;
    for (const tok of line) {
      if (tok.explanation) {
        for (const exp of tok.explanation) {
          for (const s of exp.scopes) {
            scopeSet.add(s.scopeName);
          }
        }
      }
      const color = tok.color || '#------';
      let fs = '';
      if (tok.fontStyle !== undefined) {
        if (tok.fontStyle & 1) fs = '[i]';
        else if (tok.fontStyle & 2) fs = '[b]';
        else if (tok.fontStyle & 4) fs = '[u]';
        else if (tok.fontStyle & 8) fs = '[s]';
      }
      dumpLines.push(`${color.padEnd(10)}${fs.padEnd(6)}${tok.content}`);
    }
  }

  console.log(JSON.stringify({
    name: input.name,
    lang: input.lang,
    coldMs: Math.round(coldMs * 1000) / 1000,
    warmMs: Math.round(warmMs * 1000) / 1000,
    tokens,
    scopes: scopeSet.size,
    dump: dumpLines.join('\n') + '\n',
  }));
}

h.dispose();
