"use strict";(self.webpackChunkfrontend=self.webpackChunkfrontend||[]).push([["8396"],{82822:function(e,t,i){i.d(t,{Z:()=>r});var n=i(21452),a=i(32398);let r=function(e){return"number"==typeof e||(0,a.Z)(e)&&"[object Number]"==(0,n.Z)(e)}},93953:function(e,t,i){i.d(t,{Z:()=>v});var n=i(65083),a=i(58786),r=i(56946),l=i(54470),o=i(51048),c=i(71309),d=i(60951),s=i(6445),u=i(73238);let h={name:"Flex",self:function(){return u.Z}},b=Object.assign(Object.assign({},r.Z.props),{align:String,justify:{type:String,default:"start"},inline:Boolean,vertical:Boolean,reverse:Boolean,size:{type:[String,Number,Array],default:"medium"},wrap:{type:Boolean,default:!0}}),v=(0,a.aZ)({name:"Flex",props:b,setup(e){let{mergedClsPrefixRef:t,mergedRtlRef:i}=(0,l.ZP)(e),d=(0,r.Z)("Flex","-flex",void 0,h,e,t);return{rtlEnabled:(0,o.V)("Flex",i,t),mergedClsPrefix:t,margin:(0,a.Fl)(()=>{let{size:t}=e;if(Array.isArray(t))return{horizontal:t[0],vertical:t[1]};if("number"==typeof t)return{horizontal:t,vertical:t};let{self:{[(0,c.Tl)("gap",t)]:i}}=d.value,{row:a,col:r}=(0,n.yU)(i);return{horizontal:(0,n.fQ)(r),vertical:(0,n.fQ)(a)}})}},render(){let{vertical:e,reverse:t,align:i,inline:n,justify:r,margin:l,wrap:o,mergedClsPrefix:c,rtlEnabled:u}=this,h=(0,d.x)((0,s.z)(this),!1);return h.length?(0,a.h)("div",{role:"none",class:[`${c}-flex`,u&&`${c}-flex--rtl`],style:{display:n?"inline-flex":"flex",flexDirection:e&&!t?"column":e&&t?"column-reverse":!e&&t?"row-reverse":"row",justifyContent:r,flexWrap:!o||e?"nowrap":"wrap",alignItems:i,gap:`${l.vertical}px ${l.horizontal}px`}},h):null}})},29875:function(e,t,i){let n;i.d(t,{Z:()=>$});var a=i(65083),r=i(20013),l=i(58786),o=i(76128),c=i(62594),d=i(56946),s=i(54470),u=i(32196),h=i(53198),b=i(44267),v=i(71309),f=i(93950),p=i(19595),g=i(9798),w=i(30951);let x={name:"Switch",common:g.Z,self:function(e){let{primaryColor:t,opacityDisabled:i,borderRadius:n,textColor3:a}=e;return Object.assign(Object.assign({},w.Z),{iconColor:a,textColor:"white",loadingColor:t,opacityDisabled:i,railColor:"rgba(0, 0, 0, .14)",railColorActive:t,buttonBoxShadow:"0 1px 4px 0 rgba(0, 0, 0, 0.3), inset 0 0 1px 0 rgba(0, 0, 0, 0.05)",buttonColor:"#FFF",railBorderRadiusSmall:n,railBorderRadiusMedium:n,railBorderRadiusLarge:n,buttonBorderRadiusSmall:n,buttonBorderRadiusMedium:n,buttonBorderRadiusLarge:n,boxShadowFocus:`0 0 0 2px ${(0,p.zX)(t,{alpha:.2})}`})}};var m=i(28632);let y=(0,v.cB)("switch",`
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
 `,[(0,m.c)({left:"50%",top:"50%",originalTransform:"translateX(-50%) translateY(-50%)"})]),(0,v.cE)("checked, unchecked",`
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
 `,[(0,m.c)()]),(0,v.cE)("button",`
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
 `)])]),k=Object.assign(Object.assign({},d.Z.props),{size:{type:String,default:"medium"},value:{type:[String,Number,Boolean],default:void 0},loading:Boolean,defaultValue:{type:[String,Number,Boolean],default:!1},disabled:{type:Boolean,default:void 0},round:{type:Boolean,default:!0},"onUpdate:value":[Function,Array],onUpdateValue:[Function,Array],checkedValue:{type:[String,Number,Boolean],default:!0},uncheckedValue:{type:[String,Number,Boolean],default:!1},railStyle:Function,rubberBand:{type:Boolean,default:!0},onChange:[Function,Array]}),$=(0,l.aZ)({name:"Switch",props:k,slots:Object,setup(e){void 0===n&&(n="undefined"==typeof CSS||void 0!==CSS.supports&&CSS.supports("width","max(1px)"));let{mergedClsPrefixRef:t,inlineThemeDisabled:i}=(0,s.ZP)(e),o=(0,d.Z)("Switch","-switch",y,x,e,t),c=(0,u.Z)(e),{mergedSizeRef:f,mergedDisabledRef:p}=c,g=(0,l.iH)(e.defaultValue),w=(0,l.Vh)(e,"value"),m=(0,r.Z)(w,g),k=(0,l.Fl)(()=>m.value===e.checkedValue),$=(0,l.iH)(!1),B=(0,l.iH)(!1),E=(0,l.Fl)(()=>{let{railStyle:t}=e;if(t)return t({focused:B.value,checked:k.value})});function z(t){let{"onUpdate:value":i,onChange:n,onUpdateValue:a}=e,{nTriggerFormInput:r,nTriggerFormChange:l}=c;i&&(0,b.R)(i,t),a&&(0,b.R)(a,t),n&&(0,b.R)(n,t),g.value=t,r(),l()}let S=(0,l.Fl)(()=>{let e,t,i,{value:r}=f,{self:{opacityDisabled:l,railColor:c,railColorActive:d,buttonBoxShadow:s,buttonColor:u,boxShadowFocus:h,loadingColor:b,textColor:p,iconColor:g,[(0,v.Tl)("buttonHeight",r)]:w,[(0,v.Tl)("buttonWidth",r)]:x,[(0,v.Tl)("buttonWidthPressed",r)]:m,[(0,v.Tl)("railHeight",r)]:y,[(0,v.Tl)("railWidth",r)]:k,[(0,v.Tl)("railBorderRadius",r)]:$,[(0,v.Tl)("buttonBorderRadius",r)]:B},common:{cubicBezierEaseInOut:E}}=o.value;return n?(e=`calc((${y} - ${w}) / 2)`,t=`max(${y}, ${w})`,i=`max(${k}, calc(${k} + ${w} - ${y}))`):(e=(0,a.BL)(((0,a.fQ)(y)-(0,a.fQ)(w))/2),t=(0,a.BL)(Math.max((0,a.fQ)(y),(0,a.fQ)(w))),i=(0,a.fQ)(y)>(0,a.fQ)(w)?k:(0,a.BL)((0,a.fQ)(k)+(0,a.fQ)(w)-(0,a.fQ)(y))),{"--n-bezier":E,"--n-button-border-radius":B,"--n-button-box-shadow":s,"--n-button-color":u,"--n-button-width":x,"--n-button-width-pressed":m,"--n-button-height":w,"--n-height":t,"--n-offset":e,"--n-opacity-disabled":l,"--n-rail-border-radius":$,"--n-rail-color":c,"--n-rail-color-active":d,"--n-rail-height":y,"--n-rail-width":k,"--n-width":i,"--n-box-shadow-focus":h,"--n-loading-color":b,"--n-text-color":p,"--n-icon-color":g}}),_=i?(0,h.F)("switch",(0,l.Fl)(()=>f.value[0]),S,e):void 0;return{handleClick:function(){e.loading||p.value||(m.value!==e.checkedValue?z(e.checkedValue):z(e.uncheckedValue))},handleBlur:function(){B.value=!1;let{nTriggerFormBlur:e}=c;e(),$.value=!1},handleFocus:function(){B.value=!0;let{nTriggerFormFocus:e}=c;e()},handleKeyup:function(t){e.loading||p.value||" "===t.key&&(m.value!==e.checkedValue?z(e.checkedValue):z(e.uncheckedValue),$.value=!1)},handleKeydown:function(t){e.loading||p.value||" "===t.key&&(t.preventDefault(),$.value=!0)},mergedRailStyle:E,pressed:$,mergedClsPrefix:t,mergedValue:m,checked:k,mergedDisabled:p,cssVars:i?void 0:S,themeClass:null==_?void 0:_.themeClass,onRender:null==_?void 0:_.onRender}},render(){let{mergedClsPrefix:e,mergedDisabled:t,checked:i,mergedRailStyle:n,onRender:a,$slots:r}=this;null==a||a();let{checked:d,unchecked:s,icon:u,"checked-icon":h,"unchecked-icon":b}=r,v=!((0,f.aD)(u)&&(0,f.aD)(h)&&(0,f.aD)(b));return(0,l.h)("div",{role:"switch","aria-checked":i,class:[`${e}-switch`,this.themeClass,v&&`${e}-switch--icon`,i&&`${e}-switch--active`,t&&`${e}-switch--disabled`,this.round&&`${e}-switch--round`,this.loading&&`${e}-switch--loading`,this.pressed&&`${e}-switch--pressed`,this.rubberBand&&`${e}-switch--rubber-band`],tabindex:this.mergedDisabled?void 0:0,style:this.cssVars,onClick:this.handleClick,onFocus:this.handleFocus,onBlur:this.handleBlur,onKeyup:this.handleKeyup,onKeydown:this.handleKeydown},(0,l.h)("div",{class:`${e}-switch__rail`,"aria-hidden":"true",style:n},(0,f.K9)(d,t=>(0,f.K9)(s,i=>t||i?(0,l.h)("div",{"aria-hidden":!0,class:`${e}-switch__children-placeholder`},(0,l.h)("div",{class:`${e}-switch__rail-placeholder`},(0,l.h)("div",{class:`${e}-switch__button-placeholder`}),t),(0,l.h)("div",{class:`${e}-switch__rail-placeholder`},(0,l.h)("div",{class:`${e}-switch__button-placeholder`}),i)):null)),(0,l.h)("div",{class:`${e}-switch__button`},(0,f.K9)(u,t=>(0,f.K9)(h,i=>(0,f.K9)(b,n=>(0,l.h)(o.Z,null,{default:()=>this.loading?(0,l.h)(c.Z,{key:"loading",clsPrefix:e,strokeWidth:20}):this.checked&&(i||t)?(0,l.h)("div",{class:`${e}-switch__button-icon`,key:i?"checked-icon":"icon"},i||t):!this.checked&&(n||t)?(0,l.h)("div",{class:`${e}-switch__button-icon`,key:n?"unchecked-icon":"icon"},n||t):null})))),(0,f.K9)(d,t=>t&&(0,l.h)("div",{key:"checked",class:`${e}-switch__checked`},t)),(0,f.K9)(s,t=>t&&(0,l.h)("div",{key:"unchecked",class:`${e}-switch__unchecked`},t)))))}})}}]);