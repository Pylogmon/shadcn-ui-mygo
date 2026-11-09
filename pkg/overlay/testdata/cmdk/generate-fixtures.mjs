// Optional offline oracle generation. Go consumers/tests do not require Node.
// Run: node pkg/overlay/testdata/cmdk/generate-fixtures.mjs [--check]
import { readFileSync, writeFileSync } from 'node:fs'
import { createHash } from 'node:crypto'

const read = (name) => readFileSync(new URL(name, import.meta.url), 'utf8')
const manifest = JSON.parse(read('manifest.json'))
const source = read('command-score.ts')
const digest = createHash('sha256').update(source).digest('hex')
if (digest !== manifest.files.find((file) => file.file === 'command-score.ts').sha256) {
  throw new Error('Pinned cmdk source changed; inspect the source before regenerating')
}
const signature = 'export function commandScore(string: string, abbreviation: string, aliases: string[]): number {'
if (!source.includes(signature)) throw new Error('Unexpected upstream TypeScript signature')
// Strip only the function's type annotations; the algorithm stays byte-for-byte.
const commandScore = new Function(source.replace(signature, 'function commandScore(string, abbreviation, aliases) {') + '\nreturn commandScore')()
const cases = []
const failures = []
const add = (value, query, keywords = []) => {
  try { cases.push({ value, query, keywords, score: commandScore(value, query, keywords) }) }
  catch (error) {
    if (!(error instanceof RangeError)) throw error
    failures.push({value, query, keywords, error: error.name})
  }
}
for (const [value, query, keywords] of [
  ['', ''], ['Open Project', ''], ['', 'open'], ['Open Project', 'Open Project'],
  ['Open Project', 'op'], ['Open Project', 'OP'], ['Open Project', 'OpnPrjct'],
  ['Preferences', 'prefernces'], ['Preferences', 'Prefrences'], ['Open Project', 'oen'],
  ['ouch', 'uc'], ['curtain', 'uc'], ['bad', 'bd'], ['bard', 'bd'],
  ['HTML', 'HM'], ['haml', 'HM'], ['html', 'html'], ['html5', 'html'],
  ['exist', 'eexist'], ['existing', 'exxist'], ['extension', 'etension'],
  ['Navigation', 'nvg'], ['Alpha Beta', 'ba'], ['Archive', 'zz'],
  ['foo/bar', 'fb'], ['foo_bar', 'fb'], ['foo-bar', 'fb'], ['foo bar', 'fb'],
  ['foo[bar', 'fb'], ['foo]bar', 'fb'], ['foo\\bar', 'fb'],
  ['Print document', 'paper', ['paper', 'printer']], ['Print', 'prnt', ['paper']],
  ['Print', 'paper', ['']], ['i', 'İ'], ['İstanbul', 'istanbul'], ['İstanbul', 'İ'],
  ['ΟΣ', 'ος'], ['ΟΔΟΣ', 'οδος'], ['ΣΟΣ Σ', 'σος σ'], ['ẞ Straße', 'ß'],
  ['Café', 'cf́'], ['Café', 'cfé'], ['你好世界', '你世'], ['😀 Open 🚀 Project', '😀🚀'],
  ['𐐀𐐁 test', '𐐨t'], ['a\u0085b', 'a b'], ['a\uFEFFb', 'a b'],
  ['a\u200Bb', 'a b'], ['a-b', 'a b'], ['--a', ' a'], ['one two', 'two one'],
]) add(value, query, keywords)

for (const whitespace of ['\t','\n','\v','\f','\r',' ','\u00A0','\u1680','\u2000','\u2001','\u2002','\u2003','\u2004','\u2005','\u2006','\u2007','\u2008','\u2009','\u200A','\u2028','\u2029','\u202F','\u205F','\u3000','\uFEFF']) {
  add('alpha' + whitespace + 'beta', 'a b')
  add('alpha beta', 'a' + whitespace + 'b')
}
let seed = 0x5a17c0de
const random = (n) => { seed = (Math.imul(seed, 1664525) + 1013904223) >>> 0; return seed % n }
const alphabet = Array.from('aabcehlmnoprstuvwxyzABCEHMNOPRSTU /_+.#"@[({&-\\你好世界İΟΣß😀🚀𐐀𐐨')
for (let i = 0; i < 2200; i++) {
  const chars = Array.from({length: 4 + random(28)}, () => alphabet[random(alphabet.length)])
  const value = chars.join('')
  let query
  if (i % 5 === 0) {
    query = Array.from({length: 1 + random(8)}, () => alphabet[random(alphabet.length)]).join('')
  } else {
    const take = 1 + random(Math.min(8, chars.length))
    const positions = new Set()
    while (positions.size < take) positions.add(random(chars.length))
    const selected = [...positions].sort((a,b) => a-b).map((p) => chars[p])
    if (i % 5 === 1 && selected.length > 1) [selected[0], selected[1]] = [selected[1], selected[0]]
    if (i % 5 === 2) selected.splice(random(selected.length), 0, selected[0])
    query = selected.join('')
    if (i % 5 === 3) query = query.toLowerCase()
    if (i % 5 === 4) query = query.toUpperCase()
  }
  add(value, query, i % 3 === 0 ? ['alias'+random(8), '其他 word'] : [])
}
add('a'.repeat(200)+'b', 'a'.repeat(20)+'b')
add('a '.repeat(80)+'b', 'a'.repeat(12)+'b')
const rows = (items) => items.map((item) => '    ' + JSON.stringify(item)).join(',\n')
const fixture = `{
  "sourceSHA256": ${JSON.stringify(digest)},
  "node": ${JSON.stringify(process.version)},
  "unicode": ${JSON.stringify(process.versions.unicode)},
  "cases": [
${rows(cases)}
  ],
  "failures": [
${rows(failures)}
  ]
}\n`
if (process.argv.includes('--check')) {
  if (read('scores.json') !== fixture) throw new Error('Fixtures differ from the pinned JavaScript oracle')
} else {
  writeFileSync(new URL('scores.json', import.meta.url), fixture)
}
console.log(`${cases.length} cmdk oracle scores, ${failures.length} upstream RangeError cases (${digest.slice(0,12)})`)
