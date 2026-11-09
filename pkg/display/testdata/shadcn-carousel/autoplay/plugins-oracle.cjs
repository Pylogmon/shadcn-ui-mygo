const fs=require('fs'),vm=require('vm'),crypto=require('crypto')
const engine=fs.readFileSync(process.argv[2],'utf8'),autoplay=fs.readFileSync(process.argv[3],'utf8')
const hash=x=>crypto.createHash('sha256').update(x).digest('hex')
if(hash(engine)!=='a90f9c822bd7102e4b13c985c7926f910c13fc299a368f72c9c044eb730b2507'||hash(autoplay)!=='84911d2f9b684038360e7be1019923462ffb8a34659b868f6fb046eb13edd42a')throw Error('source pins')
const context={module:{exports:{}},exports:{}};vm.createContext(context);vm.runInContext(engine+'\nglobalThis.Source={PluginsHandler,EventHandler}',context)
const lifecycle=[]
for(const name of ['active-inactive-replacement','duplicate-name','duplicate-identity','all-inactive']){
 const trace=[];const make=(id,name,active)=>({name,options:{active},init(){trace.push('init:'+id)},destroy(){trace.push('destroy:'+id)}})
 const a=make('A','alpha',true),b=make('B','beta',false),c=make('C','gamma',undefined),d=make('D','alpha',true)
 const list=name==='duplicate-name'?[a,d,c]:name==='duplicate-identity'?[a,a]:name==='all-inactive'?[b]:[a,b,c]
 const h=context.Source.PluginsHandler({optionsAtMedia:options=>options});const registry=h.init({},list)
 const map=Object.keys(registry).map(key=>({name:key,sourceIndex:list.lastIndexOf(registry[key])}))
 h.destroy();h.destroy()
 if(name==='active-inactive-replacement'){b.options.active=true;h.init({},[b,d]);h.destroy()}
 lifecycle.push({name,map,trace})
}
const ordering=[]
for(const kind of ['pointerDown','slideFocusStart'])for(const order of ['before','after']){
 const trace=[];let plugin,timer=0
 class ClockDate extends Date{constructor(...args){super(...(args.length?args:[1000]))}static now(){return 1000}}
 const ac={Date:ClockDate};vm.createContext(ac);vm.runInContext(autoplay.replace(/export \{ Autoplay as default \};?/,'globalThis.Autoplay=Autoplay'),ac)
 const state=()=>({playing:plugin.isPlaying(),remaining:plugin.timeUntilNext()})
 const handler=context.Source.EventHandler()
 const win={setTimeout(){return ++timer},clearTimeout(){}}
 const doc={visibilityState:'visible'},root={},container={}
 const api={rootNode:()=>root,containerNode:()=>container,scrollSnapList:()=>[0,.5,1],selectedScrollSnap:()=>0,internalEngine:()=>({ownerWindow:win,ownerDocument:doc,options:{watchDrag:true},eventStore:{add(){return this}}}),on:handler.on,off:handler.off,emit(type){trace.push({kind:'emit',event:type,...state()});handler.emit(type);return this}}
 handler.init(api);plugin=ac.Autoplay()
 const user=()=>trace.push({kind:'user',event:kind,...state()})
 if(order==='before')api.on(kind,user)
 plugin.init(api,{mergeOptions:(...values)=>Object.assign({},...values),optionsAtMedia:options=>options})
 if(order==='after')api.on(kind,user)
 trace.length=0;api.emit(kind);ordering.push({kind,userRegistration:order,trace})
}
const out={engineSHA256:hash(engine),autoplaySHA256:hash(autoplay),lifecycle,ordering,scope:'actual fixed PluginsHandler/EventHandler/Autoplay execution; not native bridge acceptance'}
fs.writeFileSync(process.argv[4],JSON.stringify(out,null,2)+'\n');console.log(JSON.stringify(out))
