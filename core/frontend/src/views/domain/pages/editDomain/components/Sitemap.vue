<template>
    <div class="content-wrapper">
        <n-card>
            <div class="page-tit">
                <div class="back-tool">
                    <i class="i-ri:apps-fill text-6"></i>
                </div>
                <span class="tit-content">
                    Sitemap
                </span>
            </div>

            <n-form>
                <n-form-item>
                    <div class="inline-form-items">
                        <n-input class="name"></n-input>
                        <n-input class="url"></n-input>
                        <n-select class="type"></n-select>
                        <div class="operation">
                            <i class="i-material-symbols:close-rounded text-5"></i>
                        </div>
                    </div>
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

        .inline-form-items {
            @include base.row-flex;
            width: 100%;
            justify-content: space-between;
            gap: 10px;

            .name {
                flex: 3.5;
            }

            .url {
                flex: 3.5;
            }

            .type{
                flex: 2;
            }

            .operation {
                width: 34px;
                height: 34px;
                @include base.row-flex-center;
                background: var(--color-bg-1);
                border: 1px solid var(--color-border-1);
                border-radius: 5px;
                color: var(--color-text-2);
                cursor: pointer;

                &:hover {
                    background: var(--color-bg-4);
                }
            }
        }
    }
</style>