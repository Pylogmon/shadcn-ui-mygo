const fs = require('fs'), vm = require('vm'), crypto = require('crypto')
const sourcePath = process.argv[2], outputPath = process.argv[3]
const source = fs.readFileSync(sourcePath, 'utf8')
const sha = crypto.createHash('sha256').update(source).digest('hex')
if (sha !== '84911d2f9b684038360e7be1019923462ffb8a34659b868f6fb046eb13edd42a') throw Error('wrong fixed Autoplay source')
const cases = [
 {Name:'default-controls-and-expired-ui-time',Ops:[['tick',4999],['tick',5000],['reset',5100],['stop',5200],['stop',5300],['reset',5400],['play',5500,true],['tick',12000],['play',12010,false],['tick',16010],['destroy',16100],['play',16200]]},
 {Name:'varying-delay-selected-query-vs-scheduled-deadline',Delay:[100,300,700],Ops:[['select',1050,2],['observe',1060],['tick',1100],['tick',1799],['tick',1800],['reset',1810],['tick',1910]]},
 {Name:'pointer-default-permanent-stop',Ops:[['pointerDown',1200],['pointerUp',1300],['tick',6000],['play',6100],['tick',10100]]},
 {Name:'pointer-focus-restart',Options:{stopOnInteraction:false},Ops:[['pointerDown',1200],['pointerUp',1300],['focusStart',1400],['focusOut',1500],['tick',5500]]},
 {Name:'watch-drag-off-and-focus-off',WatchDrag:false,Options:{stopOnFocusIn:false,stopOnInteraction:false},Ops:[['pointerDown',1200],['pointerUp',1300],['focusStart',1400],['focusOut',1500],['tick',5000]]},
 {Name:'hover-excludes-pointer-resume',Options:{stopOnMouseEnter:true,stopOnInteraction:false},Ops:[['mouseEnter',1200],['pointerDown',1250],['pointerUp',1300],['mouseLeave',1400],['tick',5400]]},
 {Name:'hover-permanent-stop',Options:{stopOnMouseEnter:true},Ops:[['mouseEnter',1200],['mouseLeave',1300],['tick',6000]]},
 {Name:'visibility-restarts-full-delay',Ops:[['visibility',2500,true],['tick',10000],['visibility',12000,false],['tick',15999],['tick',16000]]},
 {Name:'initially-hidden-and-manual-play',Hidden:true,Ops:[['play',1200,true],['visibility',1300,false],['tick',5300],['visibility',5400,true],['visibility',5500,false]]},
 {Name:'stop-on-last',Options:{stopOnLastSnap:true},Ops:[['tick',5000],['tick',9000],['play',9100],['tick',13100]]},
 {Name:'loop-with-jump-override',Loop:true,Options:{stopOnLastSnap:true},Ops:[['select',1100,2],['play',1200,true],['tick',5200],['tick',9200],['tick',13200]]},
 {Name:'play-on-init-false-and-one-snap',Count:1,Options:{playOnInit:false},Ops:[['observe',2000],['destroy',2100]]},
 {Name:'explicit-manual-start',Options:{playOnInit:false,jump:true},Ops:[['reset',1100],['play',1200],['tick',5200],['stop',5300],['play',5400,false],['tick',9400]]}
]

function run(spec) {
 let now=1000, selected=0, nextTimer=1
 const count=spec.Count ?? 3, loop=!!spec.Loop, timerMap=new Map(), listeners=new Map(), dom=new Map(), records=[]
 const doc={visibilityState:spec.Hidden?'hidden':'visible'}, root={}, container={}
 function state(){return {Selected:selected,Playing:plugin.isPlaying(),Remaining:plugin.timeUntilNext()}}
 function record(kind,event='',index=0,jump=false){records.push({Now:now,Kind:kind,Event:event,Index:index,Jump:jump,...state()})}
 function addDOM(target,name,fn){if(!dom.has(target))dom.set(target,new Map());const map=dom.get(target);if(!map.has(name))map.set(name,[]);map.get(name).push(fn)}
 function deliverDOM(target,name){for(const fn of (dom.get(target)?.get(name)||[]).slice())fn()}
 const win={setTimeout(fn,delay){const id=nextTimer++;timerMap.set(id,{fn,at:now+delay});return id},clearTimeout(id){timerMap.delete(id)}}
 const eventStore={add(target,name,fn){addDOM(target,name,fn);return this}}
 const api={
  rootNode:()=>root,containerNode:()=>container,scrollSnapList:()=>Array.from({length:count},(_,i)=>i/(count-1||1)),selectedScrollSnap:()=>selected,
  internalEngine:()=>({eventStore,ownerDocument:doc,ownerWindow:win,options:{watchDrag:spec.WatchDrag!==false},index:{clone:()=>({add(n){this.value=loop?(selected+n+count)%count:Math.max(0,Math.min(count-1,selected+n));return this},get(){return this.value}})}}),
  on(name,fn){if(!listeners.has(name))listeners.set(name,[]);listeners.get(name).push(fn);return this},
  off(name,fn){listeners.set(name,(listeners.get(name)||[]).filter(f=>f!==fn));return this},
  emit(name){record('event',name);for(const fn of (listeners.get(name)||[]).slice())fn();return this},
  canScrollNext:()=>loop||selected<count-1,
  scrollNext(jump){selected=loop?(selected+1)%count:Math.min(count-1,selected+1);record('scrollNext','',selected,jump)},
  scrollTo(index,jump){selected=index;record('scrollTo','',index,jump)}
 }
 class ClockDate extends Date {constructor(...args){super(...(args.length?args:[now]))}static now(){return now}}
 const context={Date:ClockDate};vm.createContext(context)
 vm.runInContext(source.replace(/export \{ Autoplay as default \};?/, 'globalThis.Autoplay = Autoplay;'),context,{filename:sourcePath})
 const opts={...(spec.Options||{})};if(spec.Delay)opts.delay=()=>spec.Delay.slice()
 const plugin=context.Autoplay(opts)
 plugin.init(api,{mergeOptions:(...v)=>Object.assign({},...v),optionsAtMedia:v=>v});record('state')
 for(const [op,at,value] of spec.Ops){
  now=at
  if(op==='tick'){
   const due=[...timerMap.entries()].filter(([,v])=>v.at<=now).sort((a,b)=>a[1].at-b[1].at)[0]
   if(due){timerMap.delete(due[0]);due[1].fn()}
  }else if(op==='select')selected=value
  else if(op==='play')plugin.play(value)
  else if(op==='stop')plugin.stop()
  else if(op==='reset')plugin.reset()
  else if(op==='destroy')plugin.destroy()
  else if(op==='visibility'){doc.visibilityState=value?'hidden':'visible';deliverDOM(doc,'visibilitychange')}
  else if(op==='mouseEnter')deliverDOM(root,'mouseenter')
  else if(op==='mouseLeave')deliverDOM(root,'mouseleave')
  else if(op==='focusOut')deliverDOM(container,'focusout')
  else if(op==='pointerDown'||op==='pointerUp'||op==='focusStart')for(const fn of (listeners.get(op==='focusStart'?'slideFocusStart':op)||[]).slice())fn()
  else if(op!=='observe')throw Error(op)
  record('state')
 }
 return {...spec,Records:records}
}
fs.writeFileSync(outputPath,JSON.stringify({SourceSHA256:sha,Cases:cases.map(run)},null,2)+'\n')
console.log(JSON.stringify({cases:cases.length,sourceSHA256:sha}))
