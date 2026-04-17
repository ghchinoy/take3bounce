import eslint from '@eslint/js';
import tseslint from 'typescript-eslint';
import eslintPluginLit from 'eslint-plugin-lit';
import globals from 'globals';

export default tseslint.config(
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  eslintPluginLit.configs['flat/recommended'] || {
    plugins: { lit: eslintPluginLit },
    rules: eslintPluginLit.configs.recommended.rules
  },
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.es2021,
      },
    },
    rules: {
      '@typescript-eslint/no-explicit-any': 'error',
      'no-console': ['warn', { allow: ['warn', 'error', 'info'] }],
    },
  },
  {
    ignores: ['dist/**', 'node_modules/**', 'vite.config.ts', '*.config.js', '*.mjs']
  }
);
