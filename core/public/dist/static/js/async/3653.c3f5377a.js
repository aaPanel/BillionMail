"use strict";(self.webpackChunkfrontend=self.webpackChunkfrontend||[]).push([["3653"],{29875:function(e,t,a){let i;a.d(t,{Z:()=>y});var l=a(65083),n=a(20013),o=a(58786),r=a(76128),c=a(62594),s=a(56946),d=a(54470),u=a(32196),h=a(53198),b=a(44267),v=a(71309),f=a(93950),p=a(19595),w=a(9798),g=a(30951);let m={name:"Switch",common:w.Z,self:function(e){let{primaryColor:t,opacityDisabled:a,borderRadius:i,textColor3:l}=e;return Object.assign(Object.assign({},g.Z),{iconColor:l,textColor:"white",loadingColor:t,opacityDisabled:a,railColor:"rgba(0, 0, 0, .14)",railColorActive:t,buttonBoxShadow:"0 1px 4px 0 rgba(0, 0, 0, 0.3), inset 0 0 1px 0 rgba(0, 0, 0, 0.05)",buttonColor:"#FFF",railBorderRadiusSmall:i,railBorderRadiusMedium:i,railBorderRadiusLarge:i,buttonBorderRadiusSmall:i,buttonBorderRadiusMedium:i,buttonBorderRadiusLarge:i,boxShadowFocus:`0 0 0 2px ${(0,p.zX)(t,{alpha:.2})}`})}};var _=a(28632);let k=(0,v.cB)("switch",`
 height: var(--n-height);
 min-width: var(--n-width);
 vertical-align: middle;
 user-select: none;
 -webkit-user-select: none;
 display: inline-flex;
 outline: none;
 justify-content: center;
 align-items: center;
`,[(0,v.cE)("children-placeholder",`
 height: var(--n-rail-height);
 display: flex;
 flex-direction: column;
 overflow: hidden;
 pointer-events: none;
 visibility: hidden;
 `),(0,v.cE)("rail-placeholder",`
 display: flex;
 flex-wrap: none;
 `),(0,v.cE)("button-placeholder",`
 width: calc(1.75 * var(--n-rail-height));
 height: var(--n-rail-height);
 `),(0,v.cB)("base-loading",`
 position: absolute;
 top: 50%;
 left: 50%;
 transform: translateX(-50%) translateY(-50%);
 font-size: calc(var(--n-button-width) - 4px);
 color: var(--n-loading-color);
 transition: color .3s var(--n-bezier);
 `,[(0,_.c)({left:"50%",top:"50%",originalTransform:"translateX(-50%) translateY(-50%)"})]),(0,v.cE)("checked, unchecked",`
 transition: color .3s var(--n-bezier);
 color: var(--n-text-color);
 box-sizing: border-box;
 position: absolute;
 white-space: nowrap;
 top: 0;
 bottom: 0;
 display: flex;
 align-items: center;
 line-height: 1;
 `),(0,v.cE)("checked",`
 right: 0;
 padding-right: calc(1.25 * var(--n-rail-height) - var(--n-offset));
 `),(0,v.cE)("unchecked",`
 left: 0;
 justify-content: flex-end;
 padding-left: calc(1.25 * var(--n-rail-height) - var(--n-offset));
 `),(0,v.c)("&:focus",[(0,v.cE)("rail",`
 box-shadow: var(--n-box-shadow-focus);
 `)]),(0,v.cM)("round",[(0,v.cE)("rail","border-radius: calc(var(--n-rail-height) / 2);",[(0,v.cE)("button","border-radius: calc(var(--n-button-height) / 2);")])]),(0,v.u4)("disabled",[(0,v.u4)("icon",[(0,v.cM)("rubber-band",[(0,v.cM)("pressed",[(0,v.cE)("rail",[(0,v.cE)("button","max-width: var(--n-button-width-pressed);")])]),(0,v.cE)("rail",[(0,v.c)("&:active",[(0,v.cE)("button","max-width: var(--n-button-width-pressed);")])]),(0,v.cM)("active",[(0,v.cM)("pressed",[(0,v.cE)("rail",[(0,v.cE)("button","left: calc(100% - var(--n-offset) - var(--n-button-width-pressed));")])]),(0,v.cE)("rail",[(0,v.c)("&:active",[(0,v.cE)("button","left: calc(100% - var(--n-offset) - var(--n-button-width-pressed));")])])])])])]),(0,v.cM)("active",[(0,v.cE)("rail",[(0,v.cE)("button","left: calc(100% - var(--n-button-width) - var(--n-offset))")])]),(0,v.cE)("rail",`
 overflow: hidden;
 height: var(--n-rail-height);
 min-width: var(--n-rail-width);
 border-radius: var(--n-rail-border-radius);
 cursor: pointer;
 position: relative;
 transition:
 opacity .3s var(--n-bezier),
 background .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier);
 background-color: var(--n-rail-color);
 `,[(0,v.cE)("button-icon",`
 color: var(--n-icon-color);
 transition: color .3s var(--n-bezier);
 font-size: calc(var(--n-button-height) - 4px);
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 display: flex;
 justify-content: center;
 align-items: center;
 line-height: 1;
 `,[(0,_.c)()]),(0,v.cE)("button",`
 align-items: center; 
 top: var(--n-offset);
 left: var(--n-offset);
 height: var(--n-button-height);
 width: var(--n-button-width-pressed);
 max-width: var(--n-button-width);
 border-radius: var(--n-button-border-radius);
 background-color: var(--n-button-color);
 box-shadow: var(--n-button-box-shadow);
 box-sizing: border-box;
 cursor: inherit;
 content: "";
 position: absolute;
 transition:
 background-color .3s var(--n-bezier),
 left .3s var(--n-bezier),
 opacity .3s var(--n-bezier),
 max-width .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier);
 `)]),(0,v.cM)("active",[(0,v.cE)("rail","background-color: var(--n-rail-color-active);")]),(0,v.cM)("loading",[(0,v.cE)("rail",`
 cursor: wait;
 `)]),(0,v.cM)("disabled",[(0,v.cE)("rail",`
 cursor: not-allowed;
 opacity: .5;
 `)])]),x=Object.assign(Object.assign({},s.Z.props),{size:{type:String,default:"medium"},value:{type:[String,Number,Boolean],default:void 0},loading:Boolean,defaultValue:{type:[String,Number,Boolean],default:!1},disabled:{type:Boolean,default:void 0},round:{type:Boolean,default:!0},"onUpdate:value":[Function,Array],onUpdateValue:[Function,Array],checkedValue:{type:[String,Number,Boolean],default:!0},uncheckedValue:{type:[String,Number,Boolean],default:!1},railStyle:Function,rubberBand:{type:Boolean,default:!0},onChange:[Function,Array]}),y=(0,o.aZ)({name:"Switch",props:x,slots:Object,setup(e){void 0===i&&(i="undefined"==typeof CSS||void 0!==CSS.supports&&CSS.supports("width","max(1px)"));let{mergedClsPrefixRef:t,inlineThemeDisabled:a}=(0,d.ZP)(e),r=(0,s.Z)("Switch","-switch",k,m,e,t),c=(0,u.Z)(e),{mergedSizeRef:f,mergedDisabledRef:p}=c,w=(0,o.iH)(e.defaultValue),g=(0,o.Vh)(e,"value"),_=(0,n.Z)(g,w),x=(0,o.Fl)(()=>_.value===e.checkedValue),y=(0,o.iH)(!1),B=(0,o.iH)(!1),$=(0,o.Fl)(()=>{let{railStyle:t}=e;if(t)return t({focused:B.value,checked:x.value})});function E(t){let{"onUpdate:value":a,onChange:i,onUpdateValue:l}=e,{nTriggerFormInput:n,nTriggerFormChange:o}=c;a&&(0,b.R)(a,t),l&&(0,b.R)(l,t),i&&(0,b.R)(i,t),w.value=t,n(),o()}let C=(0,o.Fl)(()=>{let e,t,a,{value:n}=f,{self:{opacityDisabled:o,railColor:c,railColorActive:s,buttonBoxShadow:d,buttonColor:u,boxShadowFocus:h,loadingColor:b,textColor:p,iconColor:w,[(0,v.Tl)("buttonHeight",n)]:g,[(0,v.Tl)("buttonWidth",n)]:m,[(0,v.Tl)("buttonWidthPressed",n)]:_,[(0,v.Tl)("railHeight",n)]:k,[(0,v.Tl)("railWidth",n)]:x,[(0,v.Tl)("railBorderRadius",n)]:y,[(0,v.Tl)("buttonBorderRadius",n)]:B},common:{cubicBezierEaseInOut:$}}=r.value;return i?(e=`calc((${k} - ${g}) / 2)`,t=`max(${k}, ${g})`,a=`max(${x}, calc(${x} + ${g} - ${k}))`):(e=(0,l.BL)(((0,l.fQ)(k)-(0,l.fQ)(g))/2),t=(0,l.BL)(Math.max((0,l.fQ)(k),(0,l.fQ)(g))),a=(0,l.fQ)(k)>(0,l.fQ)(g)?x:(0,l.BL)((0,l.fQ)(x)+(0,l.fQ)(g)-(0,l.fQ)(k))),{"--n-bezier":$,"--n-button-border-radius":B,"--n-button-box-shadow":d,"--n-button-color":u,"--n-button-width":m,"--n-button-width-pressed":_,"--n-button-height":g,"--n-height":t,"--n-offset":e,"--n-opacity-disabled":o,"--n-rail-border-radius":y,"--n-rail-color":c,"--n-rail-color-active":s,"--n-rail-height":k,"--n-rail-width":x,"--n-width":a,"--n-box-shadow-focus":h,"--n-loading-color":b,"--n-text-color":p,"--n-icon-color":w}}),S=a?(0,h.F)("switch",(0,o.Fl)(()=>f.value[0]),C,e):void 0;return{handleClick:function(){e.loading||p.value||(_.value!==e.checkedValue?E(e.checkedValue):E(e.uncheckedValue))},handleBlur:function(){B.value=!1;let{nTriggerFormBlur:e}=c;e(),y.value=!1},handleFocus:function(){B.value=!0;let{nTriggerFormFocus:e}=c;e()},handleKeyup:function(t){e.loading||p.value||" "===t.key&&(_.value!==e.checkedValue?E(e.checkedValue):E(e.uncheckedValue),y.value=!1)},handleKeydown:function(t){e.loading||p.value||" "===t.key&&(t.preventDefault(),y.value=!0)},mergedRailStyle:$,pressed:y,mergedClsPrefix:t,mergedValue:_,checked:x,mergedDisabled:p,cssVars:a?void 0:C,themeClass:null==S?void 0:S.themeClass,onRender:null==S?void 0:S.onRender}},render(){let{mergedClsPrefix:e,mergedDisabled:t,checked:a,mergedRailStyle:i,onRender:l,$slots:n}=this;null==l||l();let{checked:s,unchecked:d,icon:u,"checked-icon":h,"unchecked-icon":b}=n,v=!((0,f.aD)(u)&&(0,f.aD)(h)&&(0,f.aD)(b));return(0,o.h)("div",{role:"switch","aria-checked":a,class:[`${e}-switch`,this.themeClass,v&&`${e}-switch--icon`,a&&`${e}-switch--active`,t&&`${e}-switch--disabled`,this.round&&`${e}-switch--round`,this.loading&&`${e}-switch--loading`,this.pressed&&`${e}-switch--pressed`,this.rubberBand&&`${e}-switch--rubber-band`],tabindex:this.mergedDisabled?void 0:0,style:this.cssVars,onClick:this.handleClick,onFocus:this.handleFocus,onBlur:this.handleBlur,onKeyup:this.handleKeyup,onKeydown:this.handleKeydown},(0,o.h)("div",{class:`${e}-switch__rail`,"aria-hidden":"true",style:i},(0,f.K9)(s,t=>(0,f.K9)(d,a=>t||a?(0,o.h)("div",{"aria-hidden":!0,class:`${e}-switch__children-placeholder`},(0,o.h)("div",{class:`${e}-switch__rail-placeholder`},(0,o.h)("div",{class:`${e}-switch__button-placeholder`}),t),(0,o.h)("div",{class:`${e}-switch__rail-placeholder`},(0,o.h)("div",{class:`${e}-switch__button-placeholder`}),a)):null)),(0,o.h)("div",{class:`${e}-switch__button`},(0,f.K9)(u,t=>(0,f.K9)(h,a=>(0,f.K9)(b,i=>(0,o.h)(r.Z,null,{default:()=>this.loading?(0,o.h)(c.Z,{key:"loading",clsPrefix:e,strokeWidth:20}):this.checked&&(a||t)?(0,o.h)("div",{class:`${e}-switch__button-icon`,key:a?"checked-icon":"icon"},a||t):!this.checked&&(i||t)?(0,o.h)("div",{class:`${e}-switch__button-icon`,key:i?"unchecked-icon":"icon"},i||t):null})))),(0,f.K9)(s,t=>t&&(0,o.h)("div",{key:"checked",class:`${e}-switch__checked`},t)),(0,f.K9)(d,t=>t&&(0,o.h)("div",{key:"unchecked",class:`${e}-switch__unchecked`},t)))))}})},90294:function(e,t,a){a.r(t),a.d(t,{default:()=>m});var i=a(29875),l=a(57869),n=a(98048),o=a(8490),r=a(6455),c=a(73084),s=a(58786);let d={class:"content-wrapper"},u={class:"flex justify-between gap-5 items-center w-100%"},h={class:"switch-settings"},b={class:"switch-item"},v={class:"switch-item",style:{"margin-bottom":"0"}},f={class:"switch-settings"},p={class:"switch-item"},w={class:"switch-item",style:{"margin-bottom":"0"}},g=(0,s.aZ)({__name:"DomainConfiguration",setup(e){let t=(0,s.iH)([{label:"MB",value:"MB"},{label:"GB",value:"GB"},{label:"TB",value:"TB"}]);return(e,a)=>{let g=c.Z,m=r.ZP,_=o.Z,k=n.Z,x=l.ZP,y=i.Z;return(0,s.wg)(),(0,s.iD)("div",d,[(0,s.Wm)(x,null,{default:(0,s.w5)(()=>[a[3]||(a[3]=(0,s._)("div",{class:"page-tit"},[(0,s._)("div",{class:"back-tool"},[(0,s._)("i",{class:"i-cuida:mail-outline text-7"})]),(0,s._)("span",{class:"tit-content"}," Mail Domain Configuration ")],-1)),(0,s.Wm)(k,null,{default:(0,s.w5)(()=>[(0,s.Wm)(m,null,{label:(0,s.w5)(()=>a[0]||(a[0]=[(0,s._)("span",{class:"form-label"},"Domain",-1)])),default:(0,s.w5)(()=>[(0,s.Wm)(g)]),_:1}),(0,s.Wm)(m,{label:""},{label:(0,s.w5)(()=>a[1]||(a[1]=[(0,s._)("span",{class:"form-label"},"Domain Quota",-1)])),default:(0,s.w5)(()=>[(0,s._)("div",u,[(0,s.Wm)(g),(0,s.Wm)(_,{options:(0,s.SU)(t),class:"w-20"},null,8,["options"])])]),_:1}),(0,s.Wm)(m,null,{label:(0,s.w5)(()=>a[2]||(a[2]=[(0,s._)("span",{class:"form-label"},"MailBox Count",-1)])),default:(0,s.w5)(()=>[(0,s.Wm)(g)]),_:1})]),_:1})]),_:1,__:[3]}),(0,s.Wm)(x,{class:"my-5"},{default:(0,s.w5)(()=>[(0,s._)("div",h,[(0,s._)("div",b,[a[4]||(a[4]=(0,s._)("div",{class:"label"},"Track Email Opens",-1)),(0,s.Wm)(y)]),(0,s._)("div",v,[a[5]||(a[5]=(0,s._)("div",{class:"label"},"Track Link Clicks",-1)),(0,s.Wm)(y)])])]),_:1}),(0,s.Wm)(x,null,{default:(0,s.w5)(()=>[a[8]||(a[8]=(0,s._)("div",{class:"page-tit"},[(0,s._)("div",{class:"back-tool"},[(0,s._)("i",{class:"i-ri:rss-fill text-6"})]),(0,s._)("span",{class:"tit-content"}," Subscription Management ")],-1)),(0,s._)("div",f,[(0,s._)("div",p,[a[6]||(a[6]=(0,s._)("div",{class:"label"},"Include Unsubscribe Link",-1)),(0,s.Wm)(y)]),(0,s._)("div",w,[a[7]||(a[7]=(0,s._)("div",{class:"label"},"Inlcude Preferences Link",-1)),(0,s.Wm)(y)])])]),_:1,__:[8]})])}}}),m=(0,a(41748).default)(g,[["__scopeId","data-v-494f02c4"]])}}]);