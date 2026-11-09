// Run with a complete, SHA512-verified date-fns-jalali@4.1.0-0 package:
// TZ=UTC node oracle-generation.mjs /path/to/package > oracle.json
import path from 'node:path';
import {pathToFileURL} from 'node:url';
const directory=process.argv[2];
if(!directory) throw new Error('Pass the complete verified npm package directory');
const module=relative=>import(pathToFileURL(path.join(directory,relative)).href);
const {toJalali,toGregorian,j2d,isLeapJalaliYear}=await module('_lib/jalali.js');
const {addMonths,getDaysInMonth,format,getWeek,newDate}=await module('index.js');
const {faIR}=await module('locale/fa-IR.js');
const vectors=[];
for(const year of [1,2,100,1000,1399,1400,1402,1403,1404,1405,1500,275138]) {
 for(let month=1;month<=12;month++) {
  const length=j2d(year,month+1,1)-j2d(year,month,1);
  for(const day of [1,15,length]) {
   const gregorian=toGregorian(year,month,day);
   const back=toJalali(gregorian.gy,gregorian.gm,gregorian.gd);
   vectors.push({persian:[year,month,day],gregorian:[gregorian.gy,gregorian.gm,gregorian.gd],roundtrip:[back.jy,back.jm,back.jd],days:length,leap:isLeapJalaliYear(year)});
  }
 }
}
const labels=[];
for(const [y,m,d] of [[1404,3,22],[1403,12,30],[1404,1,1]]) {
 const date=newDate(y,m-1,d);
 const substitute=s=>s.replace(/[0-9]/g,c=>'۰۱۲۳۴۵۶۷۸۹'[+c]);
 const next=addMonths(date,1);const nextP=toJalali(next.getFullYear(),next.getMonth()+1,next.getDate());
 labels.push({persian:[y,m,d],caption:substitute(format(date,'LLLL y',{locale:faIR})),label:substitute(format(date,'PPPP',{locale:faIR})),day:substitute(format(date,'d',{locale:faIR})),week:substitute(String(getWeek(date,{locale:faIR}))),daysInMonth:getDaysInMonth(date),nextMonth:[nextP.jy,nextP.jm,nextP.jd]});
}
const result={source:'date-fns-jalali@4.1.0-0',timezone:'UTC',vectors,labels};
process.stdout.write(JSON.stringify(result,null,2)+'\n');
