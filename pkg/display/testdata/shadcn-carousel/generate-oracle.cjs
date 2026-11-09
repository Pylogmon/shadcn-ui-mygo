const fs=require('fs'),vm=require('vm'),crypto=require('crypto');
const source=process.argv[2]||require('path').join(__dirname,'npm/embla-carousel/cjs/embla-carousel.cjs.js');
const code=fs.readFileSync(source,'utf8');const sandbox={module:{exports:{}},exports:{}};
vm.runInNewContext(code+'\nmodule.exports.__audit={Axis,Alignment,ScrollSnaps,ScrollContain,SlideSizes,SlidesToScroll,SlideRegistry,SlideLooper,Vector1D};',sandbox,{filename:source});
const F=sandbox.module.exports.__audit;
function rect(start,size,view,mode){if(mode==='vertical'||mode==='vertical-rtl')return {left:0,right:120,width:120,top:start,bottom:start+size,height:size};if(mode==='rtl')return {left:view-start-size,right:view-start,width:size,top:0,bottom:120,height:120};return {left:start,right:start+size,width:size,top:0,bottom:120,height:120};}
function run(input){
 const {viewport,spacing,bases,align,contain,mode,loop:requestLoop}=input;const view=viewport+spacing,vertical=mode.startsWith('vertical'),axis=F.Axis(vertical?'y':'x',mode.endsWith('rtl')||mode==='rtl'?'rtl':'ltr');
 const container=rect(0,view,view,mode),sizes=bases.map(b=>b.fixed??view*b.fraction),starts=[];let total=0;const rects=sizes.map(size=>{starts.push(total);const r=rect(total,size,view,mode);total+=size;return r});
 const nodes=rects.map(()=>({style:{}}));const measured=F.SlideSizes(axis,container,rects,nodes,true,{getComputedStyle(){return {getPropertyValue(){return '0'}}}});
 const grouping=F.SlidesToScroll(axis,view,1,requestLoop,container,rects,measured.startGap,measured.endGap,2);
 const raw=F.ScrollSnaps(axis,F.Alignment(align,view),container,rects,grouping);
 const location=F.Vector1D(0);const looper=F.SlideLooper(axis,view,total,measured.slideSizes,measured.slideSizesWithGaps,raw.snaps,raw.snapsAligned,location,nodes);const loop=requestLoop&&looper.canLoop();
 const bounded=F.ScrollContain(view,total,raw.snapsAligned,contain==='none'?false:contain,2);const snaps=loop||contain==='none'?raw.snapsAligned:bounded.snapsContained;
 const registry=F.SlideRegistry(!loop&&contain!=='none',contain==='none'?false:contain,snaps,bounded.scrollContainLimit,grouping,sizes.map((_,i)=>i)).slideRegistry;
 const poses=snaps.map(snap=>{location.set(snap);const shifts=sizes.map(()=>0);if(loop)for(const p of looper.loopPoints)shifts[p.index]=p.target();return {snap,items:sizes.map((size,i)=>({index:i,outerStart:starts[i]-spacing+snap+shifts[i],outerEnd:starts[i]-spacing+snap+shifts[i]+size,contentStart:starts[i]+snap+shifts[i],contentEnd:starts[i]+snap+shifts[i]+size-spacing}))}});
 return {...input,requestedLoop:requestLoop,view,sizes,starts,total,rawSnaps:raw.snaps,alignedSnaps:raw.snapsAligned,snaps,loop,registry,poses};
}
const groups=[['full',Array(5).fill({fraction:1})],['half',Array(5).fill({fraction:.5})],['third',Array(5).fill({fraction:1/3})],['fixed',[180,240,120,300,180].map(fixed=>({fixed}))],['mixed',[.5,.25,.75,.5,.25].map(fraction=>({fraction}))],['short',Array(2).fill({fraction:.5})]];
const cases=[];for(const[name,bases]of groups)for(const spacing of[0,4,16])for(const mode of['ltr','rtl','vertical','vertical-rtl'])for(const align of['start','center','end'])for(const contain of['trimSnaps','keepSnaps','none'])for(const loop of[false,true])cases.push(run({name,viewport:600,spacing,bases,mode,align,contain,loop}));
const out={upstream:'embla-carousel@8.5.2',source,sourceSHA256:crypto.createHash('sha256').update(code).digest('hex'),functions:['Axis','Alignment','SlideSizes','SlidesToScroll','ScrollSnaps','ScrollContain','SlideRegistry','SlideLooper'],caseCount:cases.length,cases};
fs.writeFileSync(process.argv[3]||require('path').join(__dirname,'geometry-oracle-full.json'),JSON.stringify(out,null,2));console.log(JSON.stringify({cases:cases.length,sourceSHA256:out.sourceSHA256,half:cases.find(c=>c.name==='half'&&c.spacing===16&&c.mode==='ltr'&&c.align==='start'&&c.contain==='trimSnaps'&&!c.loop)?.snaps}));
