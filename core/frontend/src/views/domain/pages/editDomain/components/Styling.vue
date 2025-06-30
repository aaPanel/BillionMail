<template>
    <div class="content-wrapper">
        <n-card class="mb-5">
            <div class="page-tit">
                <div class="back-tool">
                    <i class="i-ri:apps-fill text-6"></i>
                </div>
                <span class="tit-content">
                    Color Palette
                </span>
            </div>

            <n-form>
                <n-form-item>
                    <template #label><span class="form-label">Accent Color</span></template>
                    <div class="flex justify-between gap-2.5 w-100%">
                        <n-color-picker style="width: 50px;" :modes="['hex','rgb','hsl','hsv']" :render-label="colorRenderLabel"></n-color-picker>
                        <n-input></n-input>
                    </div>
                </n-form-item>
                <n-form-item>
                    <template #label><span class="form-label">Text Color</span></template>
                    <div class="flex justify-between gap-2.5 w-100%">
                        <n-color-picker style="width: 50px;" :modes="['hex','rgb','hsl','hsv']" :render-label="colorRenderLabel"></n-color-picker>
                        <n-input></n-input>
                    </div>
                </n-form-item>
                <n-form-item>
                    <template #label><span class="form-label">Page Background</span></template>
                    <div class="flex justify-between gap-2.5 w-100%">
                        <n-color-picker style="width: 50px;" :modes="['hex','rgb','hsl','hsv']" :render-label="colorRenderLabel"></n-color-picker>
                        <n-input></n-input>
                    </div>
                </n-form-item>
                <n-form-item>
                    <template #label><span class="form-label">Container Background</span></template>
                    <div class="flex justify-between gap-2.5 w-100%">
                        <n-color-picker style="width: 50px;" :modes="['hex','rgb','hsl','hsv']" :render-label="colorRenderLabel"></n-color-picker>
                        <n-input></n-input>
                    </div>
                </n-form-item>
            </n-form>
        </n-card>

        <n-card>
            <div class="page-tit" style="margin-bottom: 10px;">
                <div class="back-tool">
                    <i class="i-ri:apps-fill text-6"></i>
                </div>
                <span class="tit-content">
                    Custom CSS
                </span>
            </div>
            <div class="sub-tit mb-5">Add custom CSS styles to your email template</div>
            <n-input type="textarea" :rows="15"></n-input>
        </n-card>
    </div>
</template>

<script setup lang="tsx">
function colorRenderLabel(){
    return ""
}
</script>

<style scoped lang="scss">
    @use "@/styles/index" as base;
    @use "./mixin.scss" as mixin;

    .content-wrapper {
        @include mixin.content-wrapper;

        .page-tit {
            @include mixin.page-tit;
        }

        .form-label {
            @include mixin.form-label;
        }

        .sub-tit {
            color: var(--color-text-1);
        }
    }
</style>