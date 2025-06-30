<template>
    <div class="content-wrapper">
        <n-card class="mb-5">
            <div class="page-tit">
                <div class="back-tool">
                    <i class="i-ri:apps-fill text-6"></i>
                </div>
                <span class="tit-content">
                    Project Details
                </span>
            </div>

            <n-form>
                <n-form-item>
                    <template #label><span class="form-label">Project Name</span></template>
                    <n-input></n-input>
                </n-form-item>
                <n-form-item>
                    <template #label><span class="form-label">Description</span></template>
                    <n-input type="textarea" :rows="7"></n-input>
                </n-form-item>
                <n-form-item>
                    <template #label><span class="form-label">Industry</span></template>
                    <n-select :options="industryOptions"></n-select>
                </n-form-item>
            </n-form>
        </n-card>

        <n-card>
            <div class="page-tit">
                <div class="back-tool">
                    <i class="i-ri:earth-fill text-6"></i>
                </div>
                <span class="tit-content">
                    Brand Logos
                </span>
            </div>

            <div class="sub-tit">
                <span class="tit">Logo System</span>
                <span>Upload your brand's logo system to maintain consistent branding across your project</span>
            </div>

            <div class="logo-operation">
                <div class="operation-item">
                    <div class="operation-tit">Primary Logo</div>
                    <div class="operation-sub-tit">Horizontal version for headers</div>
                    <div class="logo-area">
                        <div class="upload-logo-placeholder">
                            <i class="i-carbon:cloud-upload text-7"></i>
                            <span>Primary Logo</span>
                        </div>
                    </div>
                </div>
                <div class="operation-item">
                    <div class="operation-tit">Secondary Logo</div>
                    <div class="operation-sub-tit">Vertical/stacked version</div>
                    <div class="logo-area">
                        <div class="upload-logo-placeholder">
                            <i class="i-carbon:cloud-upload text-7"></i>
                            <span>Primary Logo</span>
                        </div>
                    </div>
                </div>
                <div class="operation-item">
                    <div class="operation-tit">Favicon</div>
                    <div class="operation-sub-tit">Icon for browser tabs</div>
                    <div class="logo-area">
                        <div class="upload-logo-placeholder">
                            <i class="i-carbon:cloud-upload text-7"></i>
                            <span>Primary Logo</span>
                        </div>
                    </div>
                    <div class="operation-tools">
                        <div class="tool-item">
                            <i class="i-carbon:copy text-4"></i>
                        </div>
                        <div class="tool-item">
                            <i class="i-carbon:overflow-menu-vertical text-4"></i>
                        </div>
                        <div class="tool-item">
                            <i class="i-carbon:export text-4"></i>
                        </div>
                        <div class="tool-item">
                            <i class="i-material-symbols:delete-outline text-4"></i>
                        </div>
                    </div>
                </div>
            </div>
        </n-card>
    </div>
</template>

<script setup lang="ts">
    const industryOptions = ref([
        {
            label: "test",
            value: "test"
        }
    ])
</script>

<style scoped lang="scss">
    @use "@/styles/index" as base;
    @use "./mixin.scss" as mixin;

    .content-wrapper {
        @include mixin.content-wrapper;

        .page-tit {
            @include mixin.page-tit;
        }

        .sub-tit {
            @include base.col-flex-start;
            align-items: flex-start;
            gap: 10px;
            color: var(--color-text-2);
            margin-bottom: 40px;

            .tit {
                font-size: 14px;
                font-weight: bold;
            }
        }

        .logo-operation {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 20px;

            .operation-item {
                .operation-tit {
                    margin-bottom: 5px;
                    font-size: 14px;
                    font-weight: bold;
                }

                .operation-sub-tit {
                    margin-bottom: 15px;
                }

                .logo-area {
                    width: 100%;
                    height: 130px;
                    border: 1px solid var(--color-border-1);
                    border-radius: 16px;
                    margin-bottom: 10px;

                    .upload-logo-placeholder {
                        @include base.col-flex-center;
                        gap: 5px;
                        height: 100%;
                    }
                }

                .operation-tools {
                    @include base.row-flex-start;
                    gap: 10px;
                    .tool-item {
                        width: 34px;
                        height: 34px;
                        @include base.row-flex-center;
                        background: var(--color-bg-1);
                        border:1px solid var(--color-border-1);
                        border-radius: 5px;
                        color: var(--color-text-2);
                        cursor: pointer;
                        &:hover {
                            background: var(--color-bg-4);
                        }
                    }
                }
            }
        }

        .form-label {
            @include mixin.form-label;
        }
    }
</style>