<template>
    <div class="content-wrapper">
        <n-card class="mb-5">
            <div class="page-tit" style="margin: 0;">
                <div class="back-tool">
                    <i class="i-ri:apps-fill text-6"></i>
                </div>
                <span class="tit-content">
                    AI Settings
                </span>
            </div>
        </n-card>

        <n-card>
            <div class="page-tit">
                <span class="tit-content">
                    Default AI Instructions
                </span>
            </div>

            <n-form>
                <n-form-item>
                    <template #label><span class="form-label">Instructions for Email Generation</span></template>
                    <n-input type="textarea" :rows="13"></n-input>
                </n-form-item>
            </n-form>
            <div>These instructions will be used as default guidelines when generating new emails. You can override these instructions for individual emails as needed.</div>
        </n-card>
    </div>
</template>

<script setup lang="ts">

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
    }
</style>