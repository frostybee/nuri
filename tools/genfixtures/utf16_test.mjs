import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { buildUtf16ToUtf8Map } from './generate.mjs';

describe('buildUtf16ToUtf8Map', () => {
  it('handles ASCII-only text', () => {
    const map = buildUtf16ToUtf8Map('hello');
    // Each ASCII char is 1 byte in both UTF-8 and UTF-16
    assert.deepEqual(map, [0, 1, 2, 3, 4, 5]);
  });

  it('handles CJK characters (3-byte UTF-8, 1 code unit UTF-16)', () => {
    // '中' = U+4E2D = 3 bytes in UTF-8, 1 code unit in UTF-16
    const map = buildUtf16ToUtf8Map('A中B');
    // A: utf16[0]=0, utf8=0 → map[1]=1
    // 中: utf16[1]=1, utf8=1 → map[2]=4 (3 bytes)
    // B: utf16[2]=4, utf8=4 → map[3]=5
    assert.deepEqual(map, [0, 1, 4, 5]);
  });

  it('handles emoji with surrogate pairs (4-byte UTF-8, 2 code units UTF-16)', () => {
    // '😀' = U+1F600 = 4 bytes in UTF-8, 2 code units in UTF-16 (surrogate pair)
    const map = buildUtf16ToUtf8Map('A😀B');
    // A: map[1]=1
    // 😀: map[2]=5, map[3]=5 (surrogate pair, both units map to end of 4-byte seq)
    // B: map[4]=6
    assert.deepEqual(map, [0, 1, 5, 5, 6]);
  });

  it('handles mixed content', () => {
    // '日😀x' — CJK (3 bytes, 1 unit) + emoji (4 bytes, 2 units) + ASCII (1 byte, 1 unit)
    const map = buildUtf16ToUtf8Map('日😀x');
    // 日: map[1]=3
    // 😀: map[2]=7, map[3]=7
    // x: map[4]=8
    assert.deepEqual(map, [0, 3, 7, 7, 8]);
  });

  it('handles empty string', () => {
    const map = buildUtf16ToUtf8Map('');
    assert.deepEqual(map, [0]);
  });

  it('handles 2-byte UTF-8 characters (Latin Extended)', () => {
    // 'é' = U+00E9 = 2 bytes in UTF-8, 1 code unit in UTF-16
    const map = buildUtf16ToUtf8Map('aéb');
    assert.deepEqual(map, [0, 1, 3, 4]);
  });
});
