const fs = require('fs');
const vm = require('vm');
const path = require('path');
const crypto = require('crypto');
if (process.version !== 'v26.12.0' || process.versions.icu !== '78.3') throw new Error('Use the recorded Node v26.12.0 / ICU 78.3 oracle runtime');
const root = path.resolve(__dirname, '../../../../..');
const sourcePath = 'pkg/control/testdata/shadcn-combobox/npm/base-ui-react/internals/filter.js';
const source = fs.readFileSync(path.join(root, sourcePath), 'utf8');
const context = {exports: {}, Intl, Map, require: id => {
  if (id === './resolveValueLabel') return { stringifyAsLabel: value => String(value) };
  if (id === '../utils/stringifyLocale') return { stringifyLocale: value => String(value) };
  throw new Error(id);
}};
vm.runInNewContext(source, context, {filename: sourcePath});
const cases = [
 ['en', 'Café', 'cafe'], ['en', 'Cafe\u0301', 'café'], ['en', 'café', 'cafe\u0301'],
 ['en', 'é', 'e\u0301'], ['en', 'e\u0301', 'é'], ['en', 'Ｓｗｉｆｔ', 'swift'],
 ['de', 'ä', 'a'], ['de', 'ä', 'ae'], ['de', 'a', 'ä'], ['de', 'oe', 'ö'],
 ['fr', 'œ', 'oe'], ['fr', 'é', 'e'], ['ru', 'Ёлка', 'елка'],
 ['fa', 'ک', 'ك'], ['ar', 'ك', 'ک'], ['hi', 'क़', 'क'],
 ['de', 'Straße', 'strasse'], ['de', 'weiß', 'weiss'], ['de', 'SS', 'ß'], ['en', 'ﬀ', 'ff'],
 ['en', 'co-op', 'coop!'], ['en', 'co-op', 'coop'], ['en', 'coop', 'co-op'],
 ['en', 'foo_bar', 'foobar!'], ['en', 'co op', 'coop!'], ['en', 'a+b', 'ab!'],
 ['en', '$ab', '!ab'], ['en', 'a—b', 'ab!'], ['en', 'a\tb', 'ab!'],
 ['tr', 'Isparta', 'ısparta'], ['tr', 'İzmir', 'izmir'], ['tr', 'Isparta', 'isparta'],
 ['az', 'İlham', 'ilham'], ['sv', 'Örebro', 'orebro'], ['en', 'Örebro', 'orebro'],
 ['ja', 'カタカナ', 'かたかな'], ['ja', 'ガ', 'カ'], ['ja', 'ｶﾀｶﾅ', 'カタカナ'],
 ['ja', 'ｶﾞ', 'ガ'], ['ja', 'ガ', 'ｶﾞ'], ['ko', '한글', '한글'],
 ['en', '😀', '\ufffd'], ['en', '\ufffd', '\ufffd'], ['en', 'x😀y', '😀'], ['en', 'a\u200db', 'ab!'],
 ['el', 'ΟΣ', 'ος'], ['vi', 'Hòa', 'hoa'], ['da', 'Århus', 'aarhus'],
 ['en', 'Strassek', 'strasse'], ['en', 'Go', '\ufeffGo\ufeff'], ['en', '', '\ufeff'], ['en', 'Go', '\u0085Go\u0085'],
 ['en', '...', '---'], ['en', '-', '\u200d'], ['en', '-', '\u0301'], ['en', '\u200d', '-'],
 ['en', '---', '\u200d\u200d\u200d'], ['en', '$', '-'], ['en', '--', '\u0301a'], ['en', '😀', '--'],
 ['en', '', ''], ['en', 'abc', ''], ['en', '', 'a'], ['en', 'a-b', '---']
].map(([locale,label,query]) => ({locale,label,query,contains:context.exports.getFilter({locale}).contains(label,query.trim())}));
const ariaPath = 'pkg/control/testdata/shadcn-combobox/npm/base-ui-react/combobox/root/AriaCombobox.js';
const aria = fs.readFileSync(path.join(root, ariaPath));
process.stdout.write(JSON.stringify({source:sourcePath, sourceSHA256:crypto.createHash('sha256').update(source).digest('hex'), ariaSource:ariaPath, ariaSHA256:crypto.createHash('sha256').update(aria).digest('hex'), options:{usage:'search',sensitivity:'base',ignorePunctuation:true}, labelAdapter:'String identity for plain labels; locale stub is cache-key-only', node:process.version, icu:process.versions.icu, unicode:process.versions.unicode, cldr:process.versions.cldr, cases},null,2)+'\n');
