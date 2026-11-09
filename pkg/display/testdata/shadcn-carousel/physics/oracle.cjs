const fs=require('fs'),vm=require('vm'),crypto=require('crypto');
const source=process.argv[2], code=fs.readFileSync(source,'utf8'),scope={module:{exports:{}},exports:{}};
if(crypto.createHash('sha256').update(code).digest('hex')!=='a90f9c822bd7102e4b13c985c7926f910c13fc299a368f72c9c044eb730b2507')throw Error('source pin');
vm.runInNewContext(code+'\nmodule.exports.__audit={ScrollBody,ScrollBounds,ScrollLooper,ScrollTarget,DragTracker,Animations,Vector1D,Limit,PercentOfView,Counter,mathSign,factorAbs};',scope);
const F=scope.module.exports.__audit;
function bodyCase(name,initial,duration,events,boundsConfig,loopConfig){
 const location=F.Vector1D(initial),offset=F.Vector1D(initial),previous=F.Vector1D(initial),target=F.Vector1D(initial),body=F.ScrollBody(location,offset,previous,target,duration,.68);
 const bounds=boundsConfig&&F.ScrollBounds(F.Limit(boundsConfig.min,boundsConfig.max),offset,target,body,F.PercentOfView(boundsConfig.view));
 const looper=loopConfig&&F.ScrollLooper(loopConfig.content,F.Limit(loopConfig.min,loopConfig.max),offset,[location,offset,previous,target]);
 const trace=[];for(const e of events){switch(e.action){
 case 'target':target.set(e.value);break;case 'seek':body.seek();break;
 case 'interpolate':offset.set(location.get()*e.value+previous.get()*(1-e.value));break;
 case 'duration':body.useDuration(e.value);break;case 'friction':body.useFriction(e.value);break;
 case 'baseDuration':body.useBaseDuration();break;case 'baseFriction':body.useBaseFriction();break;
 case 'offset':offset.set(e.value);break;case 'bounds':bounds.constrain(!!e.pointerDown);break;
 case 'activeBounds':bounds.toggleActive(!!e.value);break;case 'wrap':looper.loop(body.direction());break;
 }trace.push({location:location.get(),offset:offset.get(),previous:previous.get(),target:target.get(),duration:body.duration(),velocity:body.velocity(),direction:body.direction(),settled:body.settled()})}
 return{name,initial,duration,events,bounds:boundsConfig,loop:loopConfig,trace};
}
const op=(action,value,pointerDown)=>({action,...(value===undefined?{}:{value}),...(pointerDown===undefined?{}:{pointerDown})});
const ticks=(n,alpha=1,bounds,wrap)=>Array.from({length:n},()=>[...(bounds?[op('bounds',undefined,bounds==='down')]:[]),op('seek'),op('interpolate',alpha),...(wrap?[op('wrap')]:[])]).flat();
const bodies=[
 bodyCase('default-forward',0,25,[op('target',600),...ticks(100)]),
 bodyCase('default-backward',-400,25,[op('target',-1000),...ticks(100)]),
 bodyCase('instant',100,0,[op('target',-500),...ticks(2,.3)]),
 bodyCase('reverse-keeps-velocity',0,25,[op('target',600),...ticks(8),op('target',-200),...ticks(40,.6)]),
 bodyCase('duration-friction-changes',1000,25,[op('target',0),...ticks(3),op('duration',40),op('friction',.3),...ticks(8,.2),op('baseDuration'),op('baseFriction'),...ticks(40,.9)]),
 bodyCase('loop-crossing',-1100,25,[op('target',-1700),...ticks(80,.7,false,true)],null,{min:-1500,max:0,content:1500}),
 bodyCase('bounded-drag-and-release',-1100,25,[op('target',-1700),...ticks(12,.5,'down'),...ticks(80,1,'up')],{min:-1200,max:0,view:600}),
 bodyCase('offset-not-location-bounds',0,25,[op('target',200),op('offset',60),op('bounds',undefined,false),op('seek'),op('interpolate',.5),op('activeBounds',0),op('target',400),op('bounds',undefined,false),op('activeBounds',1),...ticks(40,1,'up')],{min:-1200,max:0,view:600}),
 bodyCase('drag-follow-and-release',0,25,[op('duration',.75),op('friction',.3),op('target',-70),...ticks(4,.5),op('target',-130),...ticks(3,.2),op('duration',22),op('friction',.69),op('target',-600),...ticks(80)]),
 bodyCase('settle-uses-rendered-offset',0,25,[op('target',-.0005),op('seek'),op('interpolate',0),op('seek'),op('interpolate',1)])
];
const targetInputs=[
 {current:-100,distance:-150,snap:true},{current:-100,distance:-50,snap:false},{current:-850,distance:-100,snap:true},
 {current:-850,distance:200,snap:true},{current:30,distance:100,snap:true},{current:-100,distance:-150,snap:false,loop:true},
 {current:-1600,distance:500,snap:true,loop:true},{current:400,distance:1100,snap:true,loop:true},
 {current:-600,index:0,direction:0,loop:true},{current:-600,index:0,direction:-1,loop:true},
 {current:-600,index:0,direction:1,loop:true},{current:0,index:2,direction:0,loop:true}
];
function targetFor(input){const loop=!!input.loop,snaps=[0,-300,-600,-900],content=1200,limit=F.Limit(loop?-1200:-900,0),target=F.Vector1D(input.current),api=F.ScrollTarget(loop,snaps,content,limit,target);return{api,snaps,content,limit}}
const targets=targetInputs.map(input=>{const {api,snaps,content}=targetFor(input);return{...input,snaps,content,output:input.index===undefined?api.byDistance(input.distance,input.snap):api.byIndex(input.index,input.direction)}});
const dragInputs=[
 [{point:0,time:100},{point:-60,time:150},{point:-100,time:180},{up:200}],
 [{point:0,time:100},{point:60,time:270},{point:120,time:280},{up:290}],
 [{point:0,time:100},{point:60,time:271},{point:100,time:300},{up:340}],
 [{point:0,time:100},{point:60,time:150},{up:321}],
 [{point:0,time:100},{point:10,time:150},{up:200}],
 [{point:0,time:100},{point:60,time:100},{up:100}]
];
class MouseEvent{constructor(point,time){this.clientX=point;this.timeStamp=time}}
const drags=dragInputs.map(events=>{const api=F.DragTracker({scroll:'x'},{MouseEvent}),trace=[];for(let i=0;i<events.length;i++){const e=events[i];trace.push(e.up===undefined?(i?api.pointerMove(new MouseEvent(e.point,e.time)):api.pointerDown(new MouseEvent(e.point,e.time))):api.pointerUp(new MouseEvent(0,e.up)))}return{events,trace}});
const releaseInputs=[
 {current:-40,selected:0,velocity:-.1,view:600,mouse:true},
 {current:-90,selected:0,velocity:-.7,view:600,mouse:true},
 {current:-450,selected:0,velocity:-1.2,view:600,mouse:true,skipSnaps:true},
 {current:-450,selected:0,velocity:-1.2,view:600,mouse:true},
 {current:-450,selected:1,velocity:.8,view:600,dragFree:true},
 {current:-200,selected:1,velocity:.7,view:600,dragFree:true,mouse:true},
 {current:-600,selected:2,velocity:-1.5,view:600,loop:true},
 {current:-50,selected:0,velocity:-.8,view:600,axisSign:-1},
 {current:0,selected:0,velocity:0,view:300},
 {current:-40,selected:0,velocity:-.3,view:2000,skipSnaps:true}
];
const releases=releaseInputs.map(input=>{
 const {api,snaps,content}=targetFor(input),sign=input.axisSign??1,boost=input.dragFree?(input.mouse?500:600):(input.mouse?300:400),raw=input.velocity*boost,force=sign*raw,changed=api.byDistance(0,false).index!==input.selected;
 const next=F.Counter(snaps.length-1,input.selected,!!input.loop).add(F.mathSign(force)*-1);let allowed=api.byDistance(force,!input.dragFree).distance;
 const threshold=F.Limit(50,225).constrain(input.view*.2);
 if(!input.dragFree&&Math.abs(force)>=threshold)allowed=input.skipSnaps&&changed?allowed*.5:api.byIndex(next.get(),0).distance;
 const factor=F.factorAbs(raw,allowed);
 return{...input,snaps,content,output:{target:api.byDistance(allowed,!input.dragFree),rawForce:raw,allowedForce:allowed,duration:(input.dragFree?43:25)-10*factor,friction:.68+factor/50}};
});
function scheduling(name,timestamps,resetAt,stopAt){
 const location=F.Vector1D(0),offset=F.Vector1D(0),previous=F.Vector1D(0),target=F.Vector1D(600),body=F.ScrollBody(location,offset,previous,target,25,.68);
 let queue=[],id=0,updates=0,lastAlpha=0,hiddenListener;const document={hidden:false,addEventListener(type,fn){hiddenListener=fn},removeEventListener(){}},window={requestAnimationFrame(fn){const n=++id;queue.push([n,fn]);return n},cancelAnimationFrame(n){queue=queue.filter(v=>v[0]!==n)}};
 let animation;animation=F.Animations(document,window,()=>{body.seek();updates++},alpha=>{lastAlpha=alpha;offset.set(location.get()*alpha+previous.get()*(1-alpha));if(stopAt===trace.length)animation.stop()});animation.init();const trace=[];animation.start();
 for(let i=0;i<timestamps.length;i++){if(resetAt===i){document.hidden=true;hiddenListener();document.hidden=false}const before=updates,job=queue.shift();if(job)job[1](timestamps[i]);trace.push({updates:updates-before,alpha:lastAlpha,location:location.get(),offset:offset.get(),previous:previous.get(),target:target.get(),active:queue.length>0})}
 animation.destroy();return{name,timestamps,resetAt,stopAt,trace};
}
const schedulers=[scheduling('60hz',[100,100+1000/60,100+2000/60,150,200]),scheduling('irregular',[100,105,131,160,200,201]),scheduling('zero-delta',[100,100,117,117,140]),scheduling('visibility-reset',[100,140,100000,100016.7,100035],2),scheduling('render-stops',[100,117,140,160,180],undefined,1)];
const out={sourceSHA256:crypto.createHash('sha256').update(code).digest('hex'),sourcePackage:'embla-carousel@8.5.2',bodies,targets,drags,releases,schedulers};fs.writeFileSync(process.argv[3],JSON.stringify(out,null,2)+'\n');console.log(JSON.stringify({bodies:bodies.length,targets:targets.length,drags:drags.length,releases:releases.length,schedulers:schedulers.length}));
