"use strict";(self.webpackChunkfrontend=self.webpackChunkfrontend||[]).push([["3990"],{66236:function(e,t,a){a.d(t,{W3:()=>d,ZP:()=>m});var l=a(58786),r=a(56946),n=a(54470),i=a(53198),o=a(19050),s=a(55456),c=a(71309);let u=(0,c.cB)("breadcrumb",`
 white-space: nowrap;
 cursor: default;
 line-height: var(--n-item-line-height);
`,[(0,c.c)("ul",`
 list-style: none;
 padding: 0;
 margin: 0;
 `),(0,c.c)("a",`
 color: inherit;
 text-decoration: inherit;
 `),(0,c.cB)("breadcrumb-item",`
 font-size: var(--n-font-size);
 transition: color .3s var(--n-bezier);
 display: inline-flex;
 align-items: center;
 `,[(0,c.cB)("icon",`
 font-size: 18px;
 vertical-align: -.2em;
 transition: color .3s var(--n-bezier);
 color: var(--n-item-text-color);
 `),(0,c.c)("&:not(:last-child)",[(0,c.cM)("clickable",[(0,c.cE)("link",`
 cursor: pointer;
 `,[(0,c.c)("&:hover",`
 background-color: var(--n-item-color-hover);
 `),(0,c.c)("&:active",`
 background-color: var(--n-item-color-pressed); 
 `)])])]),(0,c.cE)("link",`
 padding: 4px;
 border-radius: var(--n-item-border-radius);
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 color: var(--n-item-text-color);
 position: relative;
 `,[(0,c.c)("&:hover",`
 color: var(--n-item-text-color-hover);
 `,[(0,c.cB)("icon",`
 color: var(--n-item-text-color-hover);
 `)]),(0,c.c)("&:active",`
 color: var(--n-item-text-color-pressed);
 `,[(0,c.cB)("icon",`
 color: var(--n-item-text-color-pressed);
 `)])]),(0,c.cE)("separator",`
 margin: 0 8px;
 color: var(--n-separator-color);
 transition: color .3s var(--n-bezier);
 user-select: none;
 -webkit-user-select: none;
 `),(0,c.c)("&:last-child",[(0,c.cE)("link",`
 font-weight: var(--n-font-weight-active);
 cursor: unset;
 color: var(--n-item-text-color-active);
 `,[(0,c.cB)("icon",`
 color: var(--n-item-text-color-active);
 `)]),(0,c.cE)("separator",`
 display: none;
 `)])])]),d=(0,o.U)("n-breadcrumb"),v=Object.assign(Object.assign({},r.Z.props),{separator:{type:String,default:"/"}}),m=(0,l.aZ)({name:"Breadcrumb",props:v,setup(e){let{mergedClsPrefixRef:t,inlineThemeDisabled:a}=(0,n.ZP)(e),o=(0,r.Z)("Breadcrumb","-breadcrumb",u,s.Z,e,t);(0,l.JJ)(d,{separatorRef:(0,l.Vh)(e,"separator"),mergedClsPrefixRef:t});let c=(0,l.Fl)(()=>{let{common:{cubicBezierEaseInOut:e},self:{separatorColor:t,itemTextColor:a,itemTextColorHover:l,itemTextColorPressed:r,itemTextColorActive:n,fontSize:i,fontWeightActive:s,itemBorderRadius:c,itemColorHover:u,itemColorPressed:d,itemLineHeight:v}}=o.value;return{"--n-font-size":i,"--n-bezier":e,"--n-item-text-color":a,"--n-item-text-color-hover":l,"--n-item-text-color-pressed":r,"--n-item-text-color-active":n,"--n-separator-color":t,"--n-item-color-hover":u,"--n-item-color-pressed":d,"--n-item-border-radius":c,"--n-font-weight-active":s,"--n-item-line-height":v}}),v=a?(0,i.F)("breadcrumb",void 0,c,e):void 0;return{mergedClsPrefix:t,cssVars:a?void 0:c,themeClass:null==v?void 0:v.themeClass,onRender:null==v?void 0:v.onRender}},render(){var e;return null==(e=this.onRender)||e.call(this),(0,l.h)("nav",{class:[`${this.mergedClsPrefix}-breadcrumb`,this.themeClass],style:this.cssVars,"aria-label":"Breadcrumb"},(0,l.h)("ul",null,this.$slots))}})},9757:function(e,t,a){a.d(t,{Z:()=>o});var l=a(58786),r=a(93950),n=a(68574),i=a(66236);let o=(0,l.aZ)({name:"BreadcrumbItem",props:{separator:String,href:String,clickable:{type:Boolean,default:!0},onClick:Function},slots:Object,setup(e,{slots:t}){let a=(0,l.f3)(i.W3,null);if(!a)return()=>null;let{separatorRef:o,mergedClsPrefixRef:s}=a,c=function(e=n.j?window:null){let t=()=>{let{hash:t,host:a,hostname:l,href:r,origin:n,pathname:i,port:o,protocol:s,search:c}=(null==e?void 0:e.location)||{};return{hash:t,host:a,hostname:l,href:r,origin:n,pathname:i,port:o,protocol:s,search:c}},a=(0,l.iH)(t()),r=()=>{a.value=t()};return(0,l.bv)(()=>{e&&(e.addEventListener("popstate",r),e.addEventListener("hashchange",r))}),(0,l.SK)(()=>{e&&(e.removeEventListener("popstate",r),e.removeEventListener("hashchange",r))}),a}(),u=(0,l.Fl)(()=>e.href?"a":"span"),d=(0,l.Fl)(()=>c.value.href===e.href?"location":null);return()=>{let{value:a}=s;return(0,l.h)("li",{class:[`${a}-breadcrumb-item`,e.clickable&&`${a}-breadcrumb-item--clickable`]},(0,l.h)(u.value,{class:`${a}-breadcrumb-item__link`,"aria-current":d.value,href:e.href,onClick:e.onClick},t),(0,l.h)("span",{class:`${a}-breadcrumb-item__separator`,"aria-hidden":"true"},(0,r.gI)(t.separator,()=>{var t;return[null!=(t=e.separator)?t:o.value]})))}}})},55855:function(e,t,a){a.d(t,{Gk:()=>o,Q5:()=>v,Xr:()=>d,ZS:()=>i,_5:()=>u,gI:()=>c,jC:()=>m,ke:()=>p,mQ:()=>s,oS:()=>b});var l=a(76882),r=a(34541);let{t:n}=l.Z.global;function i(e){return r.e.get("/batch_mail/task/list",{params:e})}function o(e){return r.e.get("/batch_mail/task/stat_chart",{params:e})}function s(e){return r.e.get("/batch_mail/task/find",{params:e})}function c(e){return r.e.post("/batch_mail/task/create",e,{fetchOptions:{loading:n("market.task.loading.creating"),successMessage:!0}})}function u(e){return r.e.post("/batch_mail/task/delete",e,{fetchOptions:{loading:n("market.task.loading.deleting"),successMessage:!0}})}function d(e){return r.e.post("/batch_mail/task/pause",e,{fetchOptions:{loading:n("market.task.loading.pausing"),successMessage:!0}})}function v(e){return r.e.post("/batch_mail/task/resume",e,{fetchOptions:{loading:n("market.task.loading.resuming"),successMessage:!0}})}function m(e){return r.e.post("/batch_mail/task/send_test",e,{fetchOptions:{loading:n("market.task.loading.sendingTest"),successMessage:!0}})}function p(e){return r.e.get("/batch_mail/tracking/mail_provider",{params:e})}function b(e){return r.e.get("/batch_mail/tracking/logs",{params:e})}},6342:function(e,t,a){a.r(t),a.d(t,{default:()=>j});var l=a(57869),r=a(17977),n=a(16715),i=a(77227),o=a(58786),s=a(49468),c=a(36760),u=a(17987),d=a(68271);let v={class:"bt-time-range"},m=(0,o.aZ)({__name:"index",props:(0,o.Vf)({defaultType:{type:String,default:"today"}},{value:{},valueModifiers:{}}),emits:(0,o.Vf)(["change"],["update:value"]),setup(e,t){let{emit:a}=t,l=(0,o.tT)(e,"value"),m=(0,o.iH)(e.defaultType),p=()=>{let e=new Date,t=(0,s.i)(e);return[(0,c.b)((0,u.E)(e,-6)).getTime(),t.getTime()]},b=e=>{switch(e){case"today":l.value=(0,d.wb)();break;case"yesterday":l.value=(0,d.wb)((0,u.E)(new Date,-1));break;case"last7days":l.value=p()}a("change")},h=e=>{let t=new Date,a=(0,u.E)(t,-30);return(0,c.b)(a).getTime()>e||(0,s.i)(t).getTime()<e},f=e=>{m.value="custom",l.value=[(0,c.b)(e[0]).getTime(),(0,s.i)(e[1]).getTime()],a("change")};return b(e.defaultType),(e,t)=>{let a=i.Z,s=n.Z,c=r.Z;return(0,o.wg)(),(0,o.iD)("div",v,[(0,o.Wm)(s,{value:(0,o.SU)(m),"onUpdate:value":[t[0]||(t[0]=e=>(0,o.dq)(m)?m.value=e:null),b]},{default:(0,o.w5)(()=>[(0,o.Wm)(a,{value:"today"},{default:(0,o.w5)(()=>[(0,o.Uk)((0,o.zw)(e.$t("common.time.today")),1)]),_:1}),(0,o.Wm)(a,{value:"yesterday"},{default:(0,o.w5)(()=>[(0,o.Uk)((0,o.zw)(e.$t("common.time.yesterday")),1)]),_:1}),(0,o.Wm)(a,{value:"last7days"},{default:(0,o.w5)(()=>[(0,o.Uk)((0,o.zw)(e.$t("common.time.last7days")),1)]),_:1})]),_:1},8,["value"]),(0,o.Wm)(c,{value:l.value,type:"daterange","is-date-disabled":h,"onUpdate:value":f},null,8,["value"])])}}});var p=a(41748);let b=(0,p.default)(m,[["__scopeId","data-v-65f017db"]]);var h=a(66236),f=a(35409),g=a(9757),k=a(55855),_=a(18592),w=a(28405),y=a(94702),x=a(53475),Z=a(65965),U=a(52495);let W={class:"p-20px"},S={class:"flex justify-between items-center mb-20px"},z={class:"font-bold text-basic"},E={class:"metrics-cards"},$={class:"detail-row"},C={class:"rate-charts-card"},T=(0,o.aZ)({__name:"analytics",setup(e){let t=(0,Z.yj)(),{t:a}=(0,U.QT)(),r=(0,o.Fl)(()=>(0,d.Dx)(t.params.id||"0")),n=(0,o.iH)((0,d.wb)()),i=(0,o.iH)([]),s=(0,o.qj)({delivery:{label:a("overview.delivered"),value:0},open:{label:a("overview.opened"),value:0},click:{label:a("overview.clicked"),value:0},bounce:{label:a("overview.bounced"),value:0}}),c=(0,o.iH)({column_type:"hourly",dashboard:{delivered:0,delivery_rate:0,failed:0,failure_rate:0,sends:0},data:[]}),u=(0,o.iH)({column_type:"hourly",data:[]}),v=(0,o.iH)({column_type:"hourly",data:[]}),m=(0,o.iH)({column_type:"hourly",data:[]}),p=e=>{Object.entries(e).forEach(e=>{let[t,a]=e,l=t.replace("_rate","");l in s&&(s[l].value=a)})};async function T(){let e=await (0,k.Gk)({task_id:r.value,start_time:Math.floor(n.value[0]/1e3),end_time:Math.floor(n.value[1]/1e3)});(0,d.Kn)(e)&&(p(e.dashboard),i.value=(0,d.kJ)(e.mail_providers)?e.mail_providers:[],c.value=e.send_mail_chart,u.value=e.bounce_rate_chart,v.value=e.click_rate_chart,m.value=e.open_rate_chart)}let j=(0,o.iH)("");return(async()=>{let e=await (0,k.mQ)({id:r.value});(0,d.Kn)(e)&&(j.value=e.subject)})(),(e,t)=>{let a=(0,o.up)("router-link"),r=g.Z,d=f.ZP,p=h.ZP,k=l.ZP;return(0,o.wg)(),(0,o.iD)("div",W,[(0,o._)("div",S,[(0,o.Wm)(p,null,{default:(0,o.w5)(()=>[(0,o.Wm)(r,null,{default:(0,o.w5)(()=>[(0,o.Wm)(a,{to:"/market/task"},{default:(0,o.w5)(()=>[(0,o.Uk)((0,o.zw)(e.$t("market.task.title")),1)]),_:1})]),_:1}),(0,o.Wm)(r,null,{default:(0,o.w5)(()=>[(0,o.Uk)((0,o.zw)(e.$t("market.task.actions.analytics")),1)]),_:1}),(0,o.Wm)(r,null,{default:(0,o.w5)(()=>[(0,o.Wm)(d,{style:{"max-width":"300px"}},{default:(0,o.w5)(()=>[(0,o._)("span",z,(0,o.zw)((0,o.SU)(j)||"--"),1)]),_:1})]),_:1})]),_:1}),(0,o.Wm)(b,{value:(0,o.SU)(n),"onUpdate:value":t[0]||(t[0]=e=>(0,o.dq)(n)?n.value=e:null),"default-type":"last7days",onChange:T},null,8,["value"])]),(0,o._)("div",E,[((0,o.wg)(!0),(0,o.iD)(o.HY,null,(0,o.Ko)((0,o.SU)(s),(e,t)=>((0,o.wg)(),(0,o.j4)(_.Z,{key:t,title:e.label,value:e.value},null,8,["title","value"]))),128))]),(0,o._)("div",$,[(0,o.Wm)(k,{class:"provider-table-card",title:e.$t("overview.mailProviders")},{default:(0,o.w5)(()=>[(0,o.Wm)(w.Z,{value:(0,o.SU)(i),"onUpdate:value":t[1]||(t[1]=e=>(0,o.dq)(i)?i.value=e:null)},null,8,["value"])]),_:1},8,["title"]),(0,o.Wm)(k,{class:"send-today-card",title:e.$t("overview.sendStats")},{default:(0,o.w5)(()=>[(0,o.Wm)(y.Z,{data:(0,o.SU)(c)},null,8,["data"])]),_:1},8,["title"])]),(0,o._)("div",C,[(0,o.Wm)(x.Z,{bounce:(0,o.SU)(u),click:(0,o.SU)(v),open:(0,o.SU)(m)},null,8,["bounce","click","open"])])])}}}),j=(0,p.default)(T,[["__scopeId","data-v-04d08488"]])}}]);