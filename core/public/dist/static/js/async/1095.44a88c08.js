"use strict";(self.webpackChunkfrontend=self.webpackChunkfrontend||[]).push([["1095"],{66236:function(e,r,t){t.d(r,{W3:()=>d,ZP:()=>v});var n=t(58786),a=t(56946),o=t(54470),l=t(53198),i=t(19050),c=t(55456),s=t(71309);let u=(0,s.cB)("breadcrumb",`
 white-space: nowrap;
 cursor: default;
 line-height: var(--n-item-line-height);
`,[(0,s.c)("ul",`
 list-style: none;
 padding: 0;
 margin: 0;
 `),(0,s.c)("a",`
 color: inherit;
 text-decoration: inherit;
 `),(0,s.cB)("breadcrumb-item",`
 font-size: var(--n-font-size);
 transition: color .3s var(--n-bezier);
 display: inline-flex;
 align-items: center;
 `,[(0,s.cB)("icon",`
 font-size: 18px;
 vertical-align: -.2em;
 transition: color .3s var(--n-bezier);
 color: var(--n-item-text-color);
 `),(0,s.c)("&:not(:last-child)",[(0,s.cM)("clickable",[(0,s.cE)("link",`
 cursor: pointer;
 `,[(0,s.c)("&:hover",`
 background-color: var(--n-item-color-hover);
 `),(0,s.c)("&:active",`
 background-color: var(--n-item-color-pressed); 
 `)])])]),(0,s.cE)("link",`
 padding: 4px;
 border-radius: var(--n-item-border-radius);
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 color: var(--n-item-text-color);
 position: relative;
 `,[(0,s.c)("&:hover",`
 color: var(--n-item-text-color-hover);
 `,[(0,s.cB)("icon",`
 color: var(--n-item-text-color-hover);
 `)]),(0,s.c)("&:active",`
 color: var(--n-item-text-color-pressed);
 `,[(0,s.cB)("icon",`
 color: var(--n-item-text-color-pressed);
 `)])]),(0,s.cE)("separator",`
 margin: 0 8px;
 color: var(--n-separator-color);
 transition: color .3s var(--n-bezier);
 user-select: none;
 -webkit-user-select: none;
 `),(0,s.c)("&:last-child",[(0,s.cE)("link",`
 font-weight: var(--n-font-weight-active);
 cursor: unset;
 color: var(--n-item-text-color-active);
 `,[(0,s.cB)("icon",`
 color: var(--n-item-text-color-active);
 `)]),(0,s.cE)("separator",`
 display: none;
 `)])])]),d=(0,i.U)("n-breadcrumb"),m=Object.assign(Object.assign({},a.Z.props),{separator:{type:String,default:"/"}}),v=(0,n.aZ)({name:"Breadcrumb",props:m,setup(e){let{mergedClsPrefixRef:r,inlineThemeDisabled:t}=(0,o.ZP)(e),i=(0,a.Z)("Breadcrumb","-breadcrumb",u,c.Z,e,r);(0,n.JJ)(d,{separatorRef:(0,n.Vh)(e,"separator"),mergedClsPrefixRef:r});let s=(0,n.Fl)(()=>{let{common:{cubicBezierEaseInOut:e},self:{separatorColor:r,itemTextColor:t,itemTextColorHover:n,itemTextColorPressed:a,itemTextColorActive:o,fontSize:l,fontWeightActive:c,itemBorderRadius:s,itemColorHover:u,itemColorPressed:d,itemLineHeight:m}}=i.value;return{"--n-font-size":l,"--n-bezier":e,"--n-item-text-color":t,"--n-item-text-color-hover":n,"--n-item-text-color-pressed":a,"--n-item-text-color-active":o,"--n-separator-color":r,"--n-item-color-hover":u,"--n-item-color-pressed":d,"--n-item-border-radius":s,"--n-font-weight-active":c,"--n-item-line-height":m}}),m=t?(0,l.F)("breadcrumb",void 0,s,e):void 0;return{mergedClsPrefix:r,cssVars:t?void 0:s,themeClass:null==m?void 0:m.themeClass,onRender:null==m?void 0:m.onRender}},render(){var e;return null==(e=this.onRender)||e.call(this),(0,n.h)("nav",{class:[`${this.mergedClsPrefix}-breadcrumb`,this.themeClass],style:this.cssVars,"aria-label":"Breadcrumb"},(0,n.h)("ul",null,this.$slots))}})},9757:function(e,r,t){t.d(r,{Z:()=>i});var n=t(58786),a=t(93950),o=t(68574),l=t(66236);let i=(0,n.aZ)({name:"BreadcrumbItem",props:{separator:String,href:String,clickable:{type:Boolean,default:!0},onClick:Function},slots:Object,setup(e,{slots:r}){let t=(0,n.f3)(l.W3,null);if(!t)return()=>null;let{separatorRef:i,mergedClsPrefixRef:c}=t,s=function(e=o.j?window:null){let r=()=>{let{hash:r,host:t,hostname:n,href:a,origin:o,pathname:l,port:i,protocol:c,search:s}=(null==e?void 0:e.location)||{};return{hash:r,host:t,hostname:n,href:a,origin:o,pathname:l,port:i,protocol:c,search:s}},t=(0,n.iH)(r()),a=()=>{t.value=r()};return(0,n.bv)(()=>{e&&(e.addEventListener("popstate",a),e.addEventListener("hashchange",a))}),(0,n.SK)(()=>{e&&(e.removeEventListener("popstate",a),e.removeEventListener("hashchange",a))}),t}(),u=(0,n.Fl)(()=>e.href?"a":"span"),d=(0,n.Fl)(()=>s.value.href===e.href?"location":null);return()=>{let{value:t}=c;return(0,n.h)("li",{class:[`${t}-breadcrumb-item`,e.clickable&&`${t}-breadcrumb-item--clickable`]},(0,n.h)(u.value,{class:`${t}-breadcrumb-item__link`,"aria-current":d.value,href:e.href,onClick:e.onClick},r),(0,n.h)("span",{class:`${t}-breadcrumb-item__separator`,"aria-hidden":"true"},(0,a.gI)(r.separator,()=>{var r;return[null!=(r=e.separator)?r:i.value]})))}}})},73152:function(e,r,t){t.r(r),t.d(r,{default:()=>u});var n=t(66236),a=t(9757),o=t(58786),l=t(68271);let i={class:"h-full p-20px"},c=["src"],s=(0,o.aZ)({__name:"index",setup(e){let r=(0,o.iH)(`${window.location.origin}${l.r8?"/api":""}/rspamd/`);return(e,t)=>{let l=(0,o.up)("router-link"),s=a.Z,u=n.ZP;return(0,o.wg)(),(0,o.iD)("div",i,[(0,o.Wm)(u,{class:"mb-16px"},{default:(0,o.w5)(()=>[(0,o.Wm)(s,null,{default:(0,o.w5)(()=>[(0,o.Wm)(l,{to:"/settings"},{default:(0,o.w5)(()=>[(0,o.Uk)((0,o.zw)(e.$t("layout.menu.settings")),1)]),_:1})]),_:1}),(0,o.Wm)(s,null,{default:(0,o.w5)(()=>[(0,o.Wm)(l,{to:"/settings/service"},{default:(0,o.w5)(()=>[(0,o.Uk)((0,o.zw)(e.$t("layout.menu.service")),1)]),_:1})]),_:1}),(0,o.Wm)(s,null,{default:(0,o.w5)(()=>t[0]||(t[0]=[(0,o.Uk)("Rspamd")])),_:1,__:[0]})]),_:1}),(0,o._)("iframe",{class:"w-full",src:(0,o.SU)(r)},null,8,c)])}}}),u=(0,t(41748).default)(s,[["__scopeId","data-v-609c4574"]])}}]);