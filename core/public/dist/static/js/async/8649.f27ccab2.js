"use strict";(self.webpackChunkfrontend=self.webpackChunkfrontend||[]).push([["8649"],{23125:function(e,t,i){i.d(t,{Z:()=>n});var l=i(58786);function n(e,t){return(0,l.Fl)(()=>{for(let i of t)if(void 0!==e[i])return e[i];return e[t[t.length-1]]})}},57070:function(e,t,i){e.exports=i.p+"static/image/logo.3b07dac3.png"},60579:function(e,t,i){i.d(t,{Z:()=>h});var l=i(65083),n=i(23125),a=i(58786),s=i(62594),o=i(56946),r=i(54470),d=i(53198),c=i(71309),u=i(53573),p=i(66480);let f=(0,c.c)([(0,c.c)("@keyframes spin-rotate",`
 from {
 transform: rotate(0);
 }
 to {
 transform: rotate(360deg);
 }
 `),(0,c.cB)("spin-container",`
 position: relative;
 `,[(0,c.cB)("spin-body",`
 position: absolute;
 top: 50%;
 left: 50%;
 transform: translateX(-50%) translateY(-50%);
 `,[(0,p.h)()])]),(0,c.cB)("spin-body",`
 display: inline-flex;
 align-items: center;
 justify-content: center;
 flex-direction: column;
 `),(0,c.cB)("spin",`
 display: inline-flex;
 height: var(--n-size);
 width: var(--n-size);
 font-size: var(--n-size);
 color: var(--n-color);
 `,[(0,c.cM)("rotate",`
 animation: spin-rotate 2s linear infinite;
 `)]),(0,c.cB)("spin-description",`
 display: inline-block;
 font-size: var(--n-font-size);
 color: var(--n-text-color);
 transition: color .3s var(--n-bezier);
 margin-top: 8px;
 `),(0,c.cB)("spin-content",`
 opacity: 1;
 transition: opacity .3s var(--n-bezier);
 pointer-events: all;
 `,[(0,c.cM)("spinning",`
 user-select: none;
 -webkit-user-select: none;
 pointer-events: none;
 opacity: var(--n-opacity-spinning);
 `)])]),v={small:20,medium:18,large:16},m=Object.assign(Object.assign({},o.Z.props),{contentClass:String,contentStyle:[Object,String],description:String,stroke:String,size:{type:[String,Number],default:"medium"},show:{type:Boolean,default:!0},strokeWidth:Number,rotate:{type:Boolean,default:!0},spinning:{type:Boolean,validator:()=>!0,default:void 0},delay:Number}),h=(0,a.aZ)({name:"Spin",props:m,slots:Object,setup(e){let{mergedClsPrefixRef:t,inlineThemeDisabled:i}=(0,r.ZP)(e),s=(0,o.Z)("Spin","-spin",f,u.Z,e,t),p=(0,a.Fl)(()=>{let{size:t}=e,{common:{cubicBezierEaseInOut:i},self:n}=s.value,{opacitySpinning:a,color:o,textColor:r}=n;return{"--n-bezier":i,"--n-opacity-spinning":a,"--n-size":"number"==typeof t?(0,l.BL)(t):n[(0,c.Tl)("size",t)],"--n-color":o,"--n-text-color":r}}),m=i?(0,d.F)("spin",(0,a.Fl)(()=>{let{size:t}=e;return"number"==typeof t?String(t):t[0]}),p,e):void 0,h=(0,n.Z)(e,["spinning","show"]),g=(0,a.iH)(!1);return(0,a.m0)(t=>{let i;if(h.value){let{delay:l}=e;if(l){i=window.setTimeout(()=>{g.value=!0},l),t(()=>{clearTimeout(i)});return}}g.value=h.value}),{mergedClsPrefix:t,active:g,mergedStrokeWidth:(0,a.Fl)(()=>{let{strokeWidth:t}=e;if(void 0!==t)return t;let{size:i}=e;return v["number"==typeof i?"medium":i]}),cssVars:i?void 0:p,themeClass:null==m?void 0:m.themeClass,onRender:null==m?void 0:m.onRender}},render(){var e,t;let{$slots:i,mergedClsPrefix:l,description:n}=this,o=i.icon&&this.rotate,r=(n||i.description)&&(0,a.h)("div",{class:`${l}-spin-description`},n||(null==(e=i.description)?void 0:e.call(i))),d=i.icon?(0,a.h)("div",{class:[`${l}-spin-body`,this.themeClass]},(0,a.h)("div",{class:[`${l}-spin`,o&&`${l}-spin--rotate`],style:i.default?"":this.cssVars},i.icon()),r):(0,a.h)("div",{class:[`${l}-spin-body`,this.themeClass]},(0,a.h)(s.Z,{clsPrefix:l,style:i.default?"":this.cssVars,stroke:this.stroke,"stroke-width":this.mergedStrokeWidth,class:`${l}-spin`}),r);return null==(t=this.onRender)||t.call(this),i.default?(0,a.h)("div",{class:[`${l}-spin-container`,this.themeClass],style:this.cssVars},(0,a.h)("div",{class:[`${l}-spin-content`,this.active&&`${l}-spin-content--spinning`,this.contentClass],style:this.contentStyle},i),(0,a.h)(a.uT,{name:"fade-in-transition"},{default:()=>this.active?d:null})):d}})},24680:function(e,t,i){i.r(t),i.d(t,{default:()=>y});var l=i(98048),n=i(67655),a=i(60579),s=i(6455),o=i(73084),r=i(58786),d=i(57070),c=i(31708),u=i(68271),p=i(73282),f=i(52495),v=i(65965);let m={class:"login-container"},h={class:"login-card"},g=["src","title","alt"],w=(0,r.aZ)({__name:"index",setup(e){let{t}=(0,f.QT)(),i=(0,v.tv)(),w=(0,c.LM)(),y=(0,r.AE)("formRef"),b=(0,r.iH)(!1),S=(0,r.iH)(""),_=(0,r.iH)(!1),k=(0,r.qj)({username:"",password:"",validate_code:"",validate_code_id:""}),U={username:{required:!0,message:t("login.validation.usernameRequired"),trigger:["blur","input"]},password:{required:!0,message:t("login.validation.passwordRequired"),trigger:["blur","input"]},validate_code:{required:!0,trigger:["blur","input"],message:t("login.validation.captchaRequired")}},z=async()=>{try{_.value=!0;let e=await (0,p.o1)();(0,u.Kn)(e)&&(b.value=e.mustValidateCode,e.mustValidateCode&&(S.value=e.validateCodeBase64,k.validate_code_id=e.validateCodeId))}finally{_.value=!1}},C=(0,r.iH)(!1),Z=async()=>{try{var e;await (null==(e=y.value)?void 0:e.validate()),C.value=!0;let t=await (0,p.x4)((0,r.IU)(k));(0,u.Kn)(t)&&(w.setLoginInfo({token:t.token,refresh_token:t.refresh_token,ttl:t.ttl}),setTimeout(()=>{i.push("/")},1e3))}catch{z()}finally{C.value=!1}};return z(),(e,i)=>{let c=o.Z,u=s.ZP,p=a.Z,f=n.ZP,v=l.Z;return(0,r.wg)(),(0,r.iD)("div",m,[(0,r._)("div",h,[i[4]||(i[4]=(0,r._)("div",{class:"logo-container"},[(0,r._)("div",{class:"logo"},[(0,r._)("img",{class:"w-full",src:d})])],-1)),i[5]||(i[5]=(0,r._)("h2",{class:"login-title"},"BillionMail",-1)),(0,r.Wm)(v,{ref_key:"formRef",ref:y,size:"large",model:(0,r.SU)(k),rules:U},{default:(0,r.w5)(()=>[(0,r.Wm)(u,{"show-label":!1,path:"username"},{default:(0,r.w5)(()=>[(0,r.Wm)(c,{value:(0,r.SU)(k).username,"onUpdate:value":i[0]||(i[0]=e=>(0,r.SU)(k).username=e),placeholder:(0,r.SU)(t)("login.form.usernamePlaceholder")},null,8,["value","placeholder"])]),_:1}),(0,r.Wm)(u,{"show-label":!1,path:"password"},{default:(0,r.w5)(()=>[(0,r.Wm)(c,{value:(0,r.SU)(k).password,"onUpdate:value":i[1]||(i[1]=e=>(0,r.SU)(k).password=e),class:"password-input",type:"password","show-password-on":"click",placeholder:(0,r.SU)(t)("login.form.passwordPlaceholder"),onKeyup:(0,r.D2)(Z,["enter"])},null,8,["value","placeholder"])]),_:1}),(0,r.SU)(b)?((0,r.wg)(),(0,r.j4)(u,{key:0,"show-label":!1,path:"validate_code"},{default:(0,r.w5)(()=>[(0,r.Wm)(c,{value:(0,r.SU)(k).validate_code,"onUpdate:value":i[2]||(i[2]=e=>(0,r.SU)(k).validate_code=e),class:"flex-1",placeholder:(0,r.SU)(t)("login.form.captcha"),"input-props":{spellcheck:!1},onKeydown:(0,r.D2)(Z,["enter"])},null,8,["value","placeholder"]),(0,r.Wm)(p,{size:"small",show:(0,r.SU)(_)},{default:(0,r.w5)(()=>[(0,r._)("div",{class:"code",onClick:i[3]||(i[3]=e=>z())},[(0,r._)("img",{class:"w-full h-full",src:(0,r.SU)(S),title:(0,r.SU)(t)("login.form.changeCaptcha"),alt:(0,r.SU)(t)("login.form.captcha")},null,8,g)])]),_:1},8,["show"])]),_:1})):(0,r.kq)("",!0),(0,r.Wm)(u,{"show-label":!1,"show-feedback":!1},{default:(0,r.w5)(()=>[(0,r.Wm)(f,{type:"primary",size:"large",class:"font-bold",loading:(0,r.SU)(C),disabled:(0,r.SU)(C),block:!0,onClick:Z},{default:(0,r.w5)(()=>[(0,r.Uk)((0,r.zw)((0,r.SU)(t)("login.form.loginButton")),1)]),_:1},8,["loading","disabled"])]),_:1})]),_:1},8,["model"])])])}}}),y=(0,i(41748).default)(w,[["__scopeId","data-v-587fdef7"]])}}]);