<template>
    <div class="content-wrapper">
        <n-card class="mb-5">
            <div class="page-tit">
                <div class="back-tool">
                    <i class="i-ri:apps-fill text-6"></i>
                </div>
                <span class="tit-content">
                    Company Detail
                </span>
            </div>

            <n-form>
                <div class="flex justify-between gap-5">
                    <n-form-item class="flex-1">
                        <template #label><span class="form-label">Legal Company Name</span></template>
                        <n-input></n-input>
                    </n-form-item>
                    <n-form-item class="flex-1">
                        <template #label><span class="form-label">Website</span></template>
                        <n-input></n-input>
                    </n-form-item>
                </div>
                <n-form-item>
                    <template #label><span class="form-label">Organization Number</span></template>
                    <n-input></n-input>
                </n-form-item>
                <n-form-item>
                    <template #label><span class="form-label">Additional Company Information</span></template>
                    <n-input type="textarea" :rows="8"></n-input>
                </n-form-item>
            </n-form>
        </n-card>

        <n-card>
            <div class="page-tit">
                <div class="back-tool">
                    <i class="i-ri:apps-fill text-6"></i>
                </div>
                <span class="tit-content">
                    Contact Information
                </span>
            </div>

            <n-form>
                <n-form-item>
                    <template #label><span class="form-label">Contact Email</span></template>
                    <n-input></n-input>
                </n-form-item>
                <n-form-item>
                    <template #label><span class="form-label">Phone Number</span></template>
                    <n-input></n-input>
                </n-form-item>
                <n-form-item>
                    <template #label><span class="form-label">Support URL</span></template>
                    <n-input></n-input>
                </n-form-item>
            </n-form>
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