import { readFileSync } from 'fs'
import js from '@eslint/js'
import globals from 'globals'
import vueParser from 'vue-eslint-parser'
import vuePlugin from 'eslint-plugin-vue'
import tsParser from '@typescript-eslint/parser'
import tsPlugin from '@typescript-eslint/eslint-plugin'
import prettier from 'eslint-plugin-prettier'

const autoImport = JSON.parse(readFileSync('./.eslintrc-auto-import.json', 'utf8'))

export default [
	// Ignore vendored/generated files
	{
		ignores: ['public/static/**', 'coverage/**', 'dist/**'],
	},
	// Basic JavaScript rules
	{
		rules: {
			...js.configs.recommended.rules,
		},
	},
	// Node.js environment configuration
	{
		languageOptions: {
			globals: {
				...globals.node,
				...globals.browser,
				...globals.es2020,
				...autoImport.globals,
			},
		},
		rules: {
			'no-console': ['warn', { allow: ['warn', 'error'] }], // Disallow console
			'no-unused-vars': 'warn', // Warn on unused variables
			'no-fallthrough': 'off', // Allow fallthrough
		},
	},
	// Vue 3 configuration
	{
		files: ['**/*.vue'],
		languageOptions: {
			parser: vueParser,
			parserOptions: {
				parser: tsParser,
				ecmaVersion: 'latest',
				sourceType: 'module',
				ecmaFeatures: {
					jsx: true,
				},
			},
		},
		plugins: {
			vue: vuePlugin,
		},
		rules: {
			...vuePlugin.configs['recommended'].rules,
			'vue/multi-word-component-names': 'off', // Allow multi-word component names
			'vue/no-required-prop-with-default': 'off', // Allow required props with default values
			'vue/no-v-html': 'off',
		},
	},
	// TypeScript configuration
	{
		files: ['**/*.ts', '**/*.tsx'],
		languageOptions: {
			parser: tsParser,
		},
		plugins: {
			'@typescript-eslint': tsPlugin,
		},
		rules: {
			...tsPlugin.configs.recommended.rules,
			'@typescript-eslint/no-unused-vars': 'warn', // Warn on unused variables
			'@typescript-eslint/no-explicit-any': 'off', // Allow any type
			'@typescript-eslint/no-unsafe-function-type': 'off', // Allow unsafe function types
		},
	},
	// Prettier configuration (must be last)
	{
		plugins: {
			prettier,
		},
		rules: {
			// 'prettier/prettier': 'error',
		},
	},
]
