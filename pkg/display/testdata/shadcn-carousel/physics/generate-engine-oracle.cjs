// Executes the pinned official Engine/DragHandler, reusing the already archived
// native-rect/DOM stubs. No Go motion or translated formulas generate outputs.
const fs=require('fs'),path=require('path'),crypto=require('crypto');
const base=fs.readFileSync(path.join(__dirname,'../controller/generate-controller-oracle.cjs'),'utf8');
const source=process.argv[2]||path.join(__dirname,'../npm/embla-carousel/cjs/embla-carousel.cjs.js');
if(crypto.createHash('sha256').update(fs.readFileSync(source)).digest('hex')!=='a90f9c822bd7102e4b13c985c7926f910c13fc299a368f72c9c044eb730b2507')throw Error('source pin');
const prefix=base.slice(0,base.indexOf('const results='));
const program=String.raw`
const out={sourceSHA256:crypto.createHash('sha256').update(code).digest('hex'),sourcePackage:'embla-carousel@8.5.2',frameNS:16666666,cases:[]};
const step=1000*16666666/1e9;
function run(name,options,actions,pointer='mouse'){
 const h=fresh(undefined,{align:'start',...options},false);let clock=0;
 h.engine.animation.init();
 const doc=h.viewport.ownerDocument,win=doc.defaultView;
 win.MouseEvent=class MouseEvent{constructor(x,y){this.clientX=x;this.clientY=y;this.timeStamp=clock;this.button=0;this.buttons=1;this.target=h.viewport;this.cancelable=true}preventDefault(){}stopPropagation(){}};
 h.engine.dragHandler.init(h.api);
 function state(){let e=h.engine;return{selected:h.api.selectedScrollSnap(),previous:h.api.previousScrollSnap(),location:e.location.get(),offset:e.offsetLocation.get(),previousLocation:e.previousLocation.get(),target:e.target.get(),velocity:e.scrollBody.velocity(),duration:e.scrollBody.duration(),pointerDown:e.dragHandler.pointerDown()}}
 let trace=[];for(const type of ['scroll','select','settle','pointerDown','pointerUp'])h.api.on(type,()=>trace.push({event:type,...state()}));
 const samples=[];
 for(const action of actions){
  switch(action.kind){
  case 'next':h.api.scrollNext(!!action.jump);break;
  case 'prev':h.api.scrollPrev(!!action.jump);break;
  case 'to':h.api.scrollTo(action.index,!!action.jump);break;
  case 'frame':clock+=step;h.frame(clock);break;
  case 'idle':clock+=action.ms;break;
  case 'hide':doc.hidden=true;doc.fire('visibilitychange',{});break;
  case 'show':doc.hidden=false;h.frame(clock);break;
  case 'down':h.viewport.fire(pointer==='mouse'?'mousedown':'touchstart',pointer==='mouse'?new win.MouseEvent(action.x,action.y||60):{touches:[{clientX:action.x,clientY:action.y||60}],timeStamp:clock,target:h.viewport,cancelable:true,preventDefault(){}});break;
  case 'move':(pointer==='mouse'?doc:h.viewport).fire(pointer==='mouse'?'mousemove':'touchmove',pointer==='mouse'?new win.MouseEvent(action.x,action.y||60):{touches:[{clientX:action.x,clientY:action.y||60}],timeStamp:clock,cancelable:true,preventDefault(){}});break;
  case 'up':(pointer==='mouse'?doc:h.viewport).fire(pointer==='mouse'?'mouseup':'touchend',pointer==='mouse'?new win.MouseEvent(action.x,action.y||60):{touches:[],timeStamp:clock});break;
  }
  samples.push({action,state:state(),events:trace});trace=[];
 }
 out.cases.push({name,options,pointer,actions:samples});
}
const frames=n=>Array.from({length:n},()=>({kind:'frame'}));
run('default',{},[{kind:'next'},...frames(80)]);
run('duration40',{duration:40},[{kind:'next'},...frames(8),{kind:'to',index:3},...frames(90)]);
run('duration0',{duration:0},[{kind:'next'},...frames(2)]);
run('retarget',{},[{kind:'to',index:3},...frames(4),{kind:'to',index:0},...frames(80)]);
run('inflight-jump',{},[{kind:'next'},...frames(3),{kind:'to',index:2,jump:true},...frames(2)]);
run('hidden-reset',{},[{kind:'to',index:3},...frames(3),{kind:'hide'},{kind:'idle',ms:5000},{kind:'show'},...frames(80)]);
run('loop-prev',{loop:true},[{kind:'prev'},...frames(80)]);
for(const [name,options]of [['snap',{}],['free',{dragFree:true}],['skip',{skipSnaps:true}]])run('drag-'+name,{watchDrag:true,...options},[{kind:'down',x:450},...frames(1),{kind:'move',x:240},...frames(2),{kind:'move',x:80},...frames(1),{kind:'up',x:80},...frames(100)]);
run('drag-touch',{watchDrag:true},[{kind:'down',x:450},...frames(1),{kind:'move',x:240},...frames(2),{kind:'move',x:80},...frames(1),{kind:'up',x:80},...frames(100)],'touch');
run('drag-loop',{watchDrag:true,loop:true},[{kind:'down',x:450},...frames(1),{kind:'move',x:240},...frames(2),{kind:'move',x:80},...frames(1),{kind:'up',x:80},...frames(100)]);
fs.writeFileSync(process.argv[3]||require('path').join(__dirname,'engine-oracle.json'),JSON.stringify(out,null,2));
console.log(JSON.stringify({cases:out.cases.length,samples:out.cases.reduce((n,c)=>n+c.actions.length,0),sourceSHA256:out.sourceSHA256}));
`;
new Function('require','process','__dirname',prefix+program)(require,process,__dirname);
