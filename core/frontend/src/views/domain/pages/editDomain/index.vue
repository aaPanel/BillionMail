<template>
    <div class="wrapper">
        <!-- Page tit -->
        <div class="page-tit">
            <div class="back-tool">
                <i class="i-ci:arrow-left-lg text-4"></i>
                <div class="back-btn">back</div>
            </div>
            <span class="tit-content">
                Domain / Edit domain
            </span>
        </div>

        <!-- Content tabs -->
        <div class="content-tabs">
            <div :class="['tab-item', { active: activeTab == item }]" @click="activeTab = item"
                v-for="(item, index) in menuList" :key="index">
                <span>{{ item }}</span>
            </div>
        </div>
        <div class="dynamic-content">
            <component :is="contentMap.get(activeTab)" />
        </div>
    </div>
</template>

<script setup lang="ts">
    const DomainConfiguration = defineAsyncComponent(() => import("./components/DomainConfiguration.vue"))
    const ProjectDetails = defineAsyncComponent(() => import("./components/ProjectDetails.vue"))
    const CompanyProile = defineAsyncComponent(() => import("./components/CompanyProfile.vue"))
    const Styling = defineAsyncComponent(() => import("./components/Styling.vue"))
    const Typography = defineAsyncComponent(() => import("./components/Typography.vue"))
    const Sitemap = defineAsyncComponent(() => import("./components/Sitemap.vue"))
    const FooterSettings = defineAsyncComponent(() => import("./components/FooterSettings.vue"))
    const AISettings = defineAsyncComponent(() => import("./components/AISettings.vue"))
    const menuList = ref([
        "Domain Configuration",
        "Project Details",
        "Company Profile",
        "Styling",
        "Typography",
        "Sitemap",
        "Footer Settings",
        "AI Settings"
    ])
    const activeTab = ref("AI Settings")
    const contentMap = ref(new Map([
        ["Domain Configuration", DomainConfiguration],
        ["Project Details", ProjectDetails],
        ["Company Profile", CompanyProile],
        ["Styling", Styling],
        ["Typography", Typography],
        ["Sitemap", Sitemap],
        ["Footer Settings", FooterSettings],
        ["AI Settings", AISettings],
    ]))
</script>

<style scoped lang="scss">
    @use "@/styles/index" as base;

    .wrapper {
        display: grid;
        grid-template-rows: 40px 45px 1fr;
        height: calc(100% - 48px);
        box-sizing: border-box;
        padding: 24px;
    }

    // page title
    .page-tit {
        @include base.row-flex-start;
        margin-bottom: 20px;
        gap: 15px;

        .back-tool {
            color: var(--color-primary-1);
            @include base.row-flex-start;
            gap: 5px;
            cursor: pointer;
            @include base.transition-all;

            &:hover {
                color: var(--color-primary-3);
            }

            .back-btn {
                font-size: 18px;
            }
        }

        .tit-content {
            font-size: 18px;
            color: var(--color-text-1);
            font-weight: bold;
        }
    }

    // Content tabs
    .content-tabs {
        width: 100%;
        height: 46px;
        border-bottom: 1px solid var(--color-border-1);
        @include base.row-flex-start;

        .tab-item {
            height: 100%;
            @include base.col-flex-center;
            font-size: 14px;
            padding: 0 20px;
            cursor: pointer;
            transition: .1s color ease-in-out;

            &:hover {
                color: var(--color-primary-1);
            }

            &.active {
                border-bottom: 2px solid var(--color-primary-1);
                margin-bottom: -2px;
            }
        }
    }

    // Dynamic content
    .dynamic-content {
        height: 100%;
        box-sizing: border-box;
        padding: 20px;
    }
</style>