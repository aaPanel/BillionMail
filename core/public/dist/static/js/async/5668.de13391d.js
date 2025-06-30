"use strict";(self.webpackChunkfrontend=self.webpackChunkfrontend||[]).push([["5668"],{60309:function(e,t,a){a.d(t,{L:()=>l});function l(e,t){if(!e)return;let a=document.createElement("a");a.href=e,void 0!==t&&(a.download=t),document.body.appendChild(a),a.click(),document.body.removeChild(a)}},66236:function(e,t,a){a.d(t,{W3:()=>d,ZP:()=>p});var l=a(58786),r=a(56946),n=a(54470),o=a(53198),i=a(19050),s=a(55456),c=a(71309);let u=(0,c.cB)("breadcrumb",`
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
 `)])])]),d=(0,i.U)("n-breadcrumb"),m=Object.assign(Object.assign({},r.Z.props),{separator:{type:String,default:"/"}}),p=(0,l.aZ)({name:"Breadcrumb",props:m,setup(e){let{mergedClsPrefixRef:t,inlineThemeDisabled:a}=(0,n.ZP)(e),i=(0,r.Z)("Breadcrumb","-breadcrumb",u,s.Z,e,t);(0,l.JJ)(d,{separatorRef:(0,l.Vh)(e,"separator"),mergedClsPrefixRef:t});let c=(0,l.Fl)(()=>{let{common:{cubicBezierEaseInOut:e},self:{separatorColor:t,itemTextColor:a,itemTextColorHover:l,itemTextColorPressed:r,itemTextColorActive:n,fontSize:o,fontWeightActive:s,itemBorderRadius:c,itemColorHover:u,itemColorPressed:d,itemLineHeight:m}}=i.value;return{"--n-font-size":o,"--n-bezier":e,"--n-item-text-color":a,"--n-item-text-color-hover":l,"--n-item-text-color-pressed":r,"--n-item-text-color-active":n,"--n-separator-color":t,"--n-item-color-hover":u,"--n-item-color-pressed":d,"--n-item-border-radius":c,"--n-font-weight-active":s,"--n-item-line-height":m}}),m=a?(0,o.F)("breadcrumb",void 0,c,e):void 0;return{mergedClsPrefix:t,cssVars:a?void 0:c,themeClass:null==m?void 0:m.themeClass,onRender:null==m?void 0:m.onRender}},render(){var e;return null==(e=this.onRender)||e.call(this),(0,l.h)("nav",{class:[`${this.mergedClsPrefix}-breadcrumb`,this.themeClass],style:this.cssVars,"aria-label":"Breadcrumb"},(0,l.h)("ul",null,this.$slots))}})},9757:function(e,t,a){a.d(t,{Z:()=>i});var l=a(58786),r=a(93950),n=a(68574),o=a(66236);let i=(0,l.aZ)({name:"BreadcrumbItem",props:{separator:String,href:String,clickable:{type:Boolean,default:!0},onClick:Function},slots:Object,setup(e,{slots:t}){let a=(0,l.f3)(o.W3,null);if(!a)return()=>null;let{separatorRef:i,mergedClsPrefixRef:s}=a,c=function(e=n.j?window:null){let t=()=>{let{hash:t,host:a,hostname:l,href:r,origin:n,pathname:o,port:i,protocol:s,search:c}=(null==e?void 0:e.location)||{};return{hash:t,host:a,hostname:l,href:r,origin:n,pathname:o,port:i,protocol:s,search:c}},a=(0,l.iH)(t()),r=()=>{a.value=t()};return(0,l.bv)(()=>{e&&(e.addEventListener("popstate",r),e.addEventListener("hashchange",r))}),(0,l.SK)(()=>{e&&(e.removeEventListener("popstate",r),e.removeEventListener("hashchange",r))}),a}(),u=(0,l.Fl)(()=>e.href?"a":"span"),d=(0,l.Fl)(()=>c.value.href===e.href?"location":null);return()=>{let{value:a}=s;return(0,l.h)("li",{class:[`${a}-breadcrumb-item`,e.clickable&&`${a}-breadcrumb-item--clickable`]},(0,l.h)(u.value,{class:`${a}-breadcrumb-item__link`,"aria-current":d.value,href:e.href,onClick:e.onClick},t),(0,l.h)("span",{class:`${a}-breadcrumb-item__separator`,"aria-hidden":"true"},(0,r.gI)(t.separator,()=>{var t;return[null!=(t=e.separator)?t:i.value]})))}}})},50144:function(e,t,a){a.d(t,{Z:()=>c});var l=a(58786),r=a(56946),n=a(54470),o=a(49942),i=a(44108);let s=Object.assign(Object.assign({},o.Kd),r.Z.props),c=(0,l.aZ)({name:"Tooltip",props:s,slots:Object,__popover__:!0,setup(e){let{mergedClsPrefixRef:t}=(0,n.ZP)(e),a=(0,r.Z)("Tooltip","-tooltip",void 0,i.Z,e,t),o=(0,l.iH)(null);return Object.assign(Object.assign({},{syncPosition(){o.value.syncPosition()},setShow(e){o.value.setShow(e)}}),{popoverRef:o,mergedTheme:a,popoverThemeOverrides:(0,l.Fl)(()=>a.value.self)})},render(){let{mergedTheme:e,internalExtraClass:t}=this;return(0,l.h)(o.ZP,Object.assign(Object.assign({},this.$props),{theme:e.peers.Popover,themeOverrides:e.peerOverrides.Popover,builtinThemeOverrides:this.popoverThemeOverrides,internalExtraClass:t.concat("tooltip"),ref:"popoverRef"}),this.$slots)}})},31651:function(e,t,a){a.d(t,{A8:()=>o,NI:()=>i,QB:()=>u,WF:()=>c,Xc:()=>m,ef:()=>d,ij:()=>s});var l=a(76882),r=a(34541);let{t:n}=l.Z.global;function o(e){return r.e.get("/email_template/list",{params:e})}function i(){return r.e.get("/email_template/all")}function s(e){return r.e.get("/email_template/get",{params:e})}function c(e){return r.e.post("/email_template/create",e,{fetchOptions:{loading:n("market.template.loading.creating"),successMessage:!0}})}function u(e){return r.e.post("/email_template/update",e,{fetchOptions:{loading:n("market.template.loading.updating"),successMessage:!0}})}function d(e){return r.e.post("/email_template/copy",e,{fetchOptions:{loading:n("market.template.loading.copying"),successMessage:!0}})}function m(e){return r.e.post("/email_template/delete",e,{fetchOptions:{loading:n("market.template.loading.deleting"),successMessage:!0}})}},88253:function(e,t,a){a.d(t,{Z:()=>i});var l=a(60579),r=a(58786),n=a(6768);let o=(0,r.aZ)({__name:"index",props:(0,r.Vf)({language:{type:String,default:"html"},theme:{type:String,default:"vs-dark"},options:{type:Object,default:()=>({})}},{value:{default:""},valueModifiers:{}}),emits:(0,r.Vf)(["save"],["update:value"]),setup(e,t){let{expose:a,emit:o}=t,i=(0,r.tT)(e,"value"),s=(0,r.iH)(null),c=null,u=(0,r.iH)(!1);n.Z.config({paths:{vs:`${location.origin}/static/plugin/monaco`}});let d=async()=>{if(s.value&&!c)try{u.value=!0;let t=await n.Z.init();(c=t.editor.create(s.value,{value:i.value,language:e.language,theme:e.theme,minimap:{enabled:!1},automaticLayout:!0,scrollBeyondLastLine:!1,fontSize:14,tabSize:2,wordWrap:"on",...e.options})).onDidChangeModelContent(()=>{i.value=(null==c?void 0:c.getValue())||""}),c.addCommand(t.KeyMod.CtrlCmd|t.KeyCode.KeyS,function(){o("save")})}finally{u.value=!1}},m=()=>{c&&(c.dispose(),c=null)};return(0,r.YP)(()=>i.value,e=>{c&&e!==c.getValue()&&c.setValue(e)}),(0,r.bv)(()=>{d()}),(0,r.Jd)(()=>{m()}),a({getEditor:()=>c}),(e,t)=>{let a=l.Z;return(0,r.wg)(),(0,r.j4)(a,{class:"h-full",show:(0,r.SU)(u)},{default:(0,r.w5)(()=>[(0,r._)("div",{ref_key:"containerRef",ref:s,class:"monaco-editor-container"},null,512)]),_:1},8,["show"])}}}),i=(0,a(41748).default)(o,[["__scopeId","data-v-72ffe419"]])},53472:function(e,t,a){a.d(t,{Z:()=>n});var l=a(58786);let r=["height","srcdoc"],n=(0,l.aZ)({__name:"index",props:{value:{type:String,default:""},height:{type:[String,Number],default:"100%"}},setup:e=>(t,a)=>((0,l.wg)(),(0,l.iD)("iframe",{width:"100%",height:e.height,srcdoc:e.value,style:{border:"none",borderRadius:"4px"}},null,8,r))})},80637:function(e,t,a){a.r(t),a.d(t,{default:()=>O});var l=a(67655),r=a(46244),n=a(57869),o=a(53472),i=a(88253),s=a(3179),c=a(6455),u=a(73084),d=a(66236),m=a(9757),p=a(58786),v=a(68271),f=a(31651),h=a(65965);let g={class:"template-edit"},b={class:"w-400px"},_={class:"editor-area"},w={class:"mb-16px"},k={class:"flex-1 overflow-hidden"},y={class:"preview-title"},Z={class:"flex-1 overflow-hidden"},x={class:"action-buttons"},C=(0,p.aZ)({__name:"edit",setup(e){let t=(0,h.yj)(),a=(0,p.iH)(`${t.params.id}`),C=(0,p.qj)({id:0,temp_name:"",content:""}),O=e=>{let t=new FileReader;t.onload=e=>{var t;C.content=null==(t=e.target)?void 0:t.result},e.file&&t.readAsText(e.file)},S=async()=>{},z=()=>{window.history.back()};return(async()=>{let e=await (0,f.ij)({id:a.value});(0,v.Kn)(e)&&(C.id=e.id,C.temp_name=e.temp_name)})(),(e,t)=>{let a=(0,p.up)("router-link"),v=m.Z,f=d.ZP,h=u.Z,W=c.ZP,j=s.Z,$=i.Z,P=o.Z,U=n.ZP,B=r.Z,E=l.ZP;return(0,p.wg)(),(0,p.iD)("div",g,[(0,p.Wm)(f,{class:"mb-20px"},{default:(0,p.w5)(()=>[(0,p.Wm)(v,null,{default:(0,p.w5)(()=>[(0,p.Wm)(a,{to:"/market/template"},{default:(0,p.w5)(()=>[(0,p.Uk)((0,p.zw)(e.$t("market.template.title")),1)]),_:1})]),_:1}),(0,p.Wm)(v,null,{default:(0,p.w5)(()=>[(0,p.Uk)((0,p.zw)((0,p.SU)(C).temp_name),1)]),_:1})]),_:1}),(0,p.Wm)(B,{class:"template-form","label-placement":"top"},{default:(0,p.w5)(()=>[(0,p.Wm)(W,{label:e.$t("market.template.name")},{default:(0,p.w5)(()=>[(0,p._)("div",b,[(0,p.Wm)(h,{value:(0,p.SU)(C).temp_name,"onUpdate:value":t[0]||(t[0]=e=>(0,p.SU)(C).temp_name=e)},null,8,["value"])])]),_:1},8,["label"]),(0,p.Wm)(W,{class:"editor-form-item",label:e.$t("market.template.content"),"show-feedback":!1},{default:(0,p.w5)(()=>[(0,p._)("div",_,[(0,p._)("div",w,[(0,p.Wm)(j,{mode:"button","is-upload":!1,accept:["html"],onChange:O},{default:(0,p.w5)(()=>[(0,p.Uk)((0,p.zw)(e.$t("market.template.uploadHtml")),1)]),_:1})]),(0,p._)("div",k,[(0,p.Wm)($,{value:(0,p.SU)(C).content,"onUpdate:value":t[1]||(t[1]=e=>(0,p.SU)(C).content=e),language:"html"},null,8,["value"])])]),(0,p.Wm)(U,{class:"preview-area","content-style":{display:"flex",flexDirection:"column"}},{default:(0,p.w5)(()=>[(0,p._)("div",y,(0,p.zw)(e.$t("common.actions.preview")),1),(0,p._)("div",Z,[(0,p.Wm)(P,{value:(0,p.SU)(C).content},null,8,["value"])])]),_:1})]),_:1},8,["label"])]),_:1}),(0,p._)("div",x,[(0,p.Wm)(E,{type:"primary",onClick:S},{default:(0,p.w5)(()=>[(0,p.Uk)((0,p.zw)(e.$t("common.actions.save")),1)]),_:1}),(0,p.Wm)(E,{class:"ml-16px",secondary:"",onClick:z},{default:(0,p.w5)(()=>[(0,p.Uk)((0,p.zw)(e.$t("common.actions.back")),1)]),_:1})])])}}}),O=(0,a(41748).default)(C,[["__scopeId","data-v-32d33acc"]])}}]);