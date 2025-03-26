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
			'no-console': 'warn', // Disallow console
			'no-unused-vars': 'warn', // Warn on unused variables
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
		},
	},
	// Prettier configuration (must be last)
	{
		plugins: {
			prettier,
		},
		rules: {
			'prettier/prettier': 'error',
		},
	},
]
