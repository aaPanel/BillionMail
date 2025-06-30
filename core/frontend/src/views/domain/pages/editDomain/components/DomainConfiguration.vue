<template>
    <div class="content-wrapper">
        <n-card>
            <!-- title -->
            <div class="page-tit">
                <div class="back-tool">
                    <i class="i-cuida:mail-outline text-7"></i>
                </div>
                <span class="tit-content">
                    Mail Domain Configuration
                </span>
            </div>

            <!-- form data -->
            <n-form>
                <n-form-item>
                    <template #label><span class="form-label">Domain</span></template>
                    <n-input></n-input>
                </n-form-item>
                <n-form-item label="">
                    <template #label><span class="form-label">Domain Quota</span></template>
                    <div class="flex justify-between gap-5 items-center w-100%">
                        <n-input></n-input>
                        <n-select :options="uinitOptions" class="w-20"></n-select>
                    </div>
                </n-form-item>
                <n-form-item>
                    <template #label><span class="form-label">MailBox Count</span></template>
                    <n-input></n-input>
                </n-form-item>
            </n-form>
        </n-card>

        <n-card class="my-5">
            <div class="switch-settings">
                <div class="switch-item">
                    <div class="label">Track Email Opens</div>
                    <n-switch></n-switch>
                </div>
                <div class="switch-item" style="margin-bottom: 0;">
                    <div class="label">Track Link Clicks</div>
                    <n-switch></n-switch>
                </div>
            </div>
        </n-card>

        <n-card>
            <!-- title -->
            <div class="page-tit">
                <div class="back-tool">
                    <i class="i-ri:rss-fill text-6"></i>
                </div>
                <span class="tit-content">
                    Subscription Management
                </span>
            </div>

            <div class="switch-settings">
                <div class="switch-item">
                    <div class="label">Include Unsubscribe Link</div>
                    <n-switch></n-switch>
                </div>
                <div class="switch-item" style="margin-bottom: 0;">
                    <div class="label">Inlcude Preferences Link</div>
                    <n-switch></n-switch>
                </div>
            </div>
        </n-card>
    </div>
</template>

<script setup lang="ts">
    const uinitOptions = ref([
        {
            label: "MB",
            value: "MB"
        },
        {
            label: "GB",
            value: "GB"
        },
        {
            label: "TB",
            value: "TB"
        },
    ])
</script>

<style scoped lang="scss">
    @use "@/styles/index" as base;
    @use "./mixin.scss" as mixin;

    .content-wrapper {
        @include mixin.content-wrapper;
        // title
        .page-tit {
            @include mixin.page-tit;
        }

        .form-label {
            @include mixin.form-label;
        }

        // Switch settings
        .switch-settings {
            @mixin settings-label {
                font-size: 14px;
                font-weight: bold;
            }

            .switch-item {
                @include base.row-flex;
                justify-content: space-between;
                margin-bottom: 15px;

                .label {
                    color: var(--color-text-2);
                    @include settings-label()
                }
            }
        }
    }


</style>