import AutoImport from 'unplugin-auto-import/rspack'
import Components from 'unplugin-vue-components/rspack'
import { defineConfig } from '@rsbuild/core'
import { pluginBabel } from '@rsbuild/plugin-babel'
import { pluginVue } from '@rsbuild/plugin-vue'
import { pluginVueJsx } from '@rsbuild/plugin-vue-jsx'
import { pluginSass } from '@rsbuild/plugin-sass'
import { pluginEslint } from '@rsbuild/plugin-eslint'
import { UnoCSSRspackPlugin } from '@unocss/webpack/rspack'
import { NaiveUiResolver } from 'unplugin-vue-components/resolvers'
import { pluginBasicSsl } from '@rsbuild/plugin-basic-ssl'
import { getEnv, getServer } from './build/utils'

const server = getServer()

export default defineConfig({
	plugins: [
		pluginBabel({
			include: /\.(?:jsx|tsx)$/,
		}),
		pluginVue(),
		pluginVueJsx(),
		pluginSass(),
		...(server.https ? [pluginBasicSsl()] : []),
		pluginEslint({
			eslintPluginOptions: {
				cwd: __dirname,
				configType: 'flat',
				extensions: ['.js', '.jsx', '.ts', '.tsx', '.vue'],
			},
		}),
	],
	tools: {
		rspack: {
			plugins: [
				UnoCSSRspackPlugin({
					configFile: './uno.config.ts',
				}),
				AutoImport({
					dts: 'types/auto-import.d.ts',
					imports: ['vue', 'vue-router', 'vue-i18n'],
					eslintrc: {
						enabled: true,
						filepath: '.eslintrc-auto-import.json',
					},
				}),
				Components({
					dts: 'types/components.d.ts',
					dirs: ['src/components/**/*'],
					extensions: ['vue', 'tsx'],
					resolvers: [NaiveUiResolver()],
				}),
			],
		},
	},
	html: {
		template: './index.html',
	},
	resolve: {
		extensions: ['.ts', '.tsx', '.mjs', '.js', '.jsx', '.json', '.d.ts'],
		alias: {
			'@': './src/',
			'@images': './src/assets/images/',
		},
	},
	source: {
		define: {
			'import.meta.env': JSON.stringify({
				SERVER: server,
				API_URL_PREFIX: getEnv('API_URL_PREFIX'),
			}),
		},
	},
	server: {
		proxy: {
			'/api': {
				target: server.address,
				secure: false,
				changeOrigin: true,
				pathRewrite: { '^/api': '' },
			},
		},
	},
})
