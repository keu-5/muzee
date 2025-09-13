import { fixupConfigRules, fixupPluginRules } from "@eslint/compat";
import { FlatCompat } from "@eslint/eslintrc";
import js from "@eslint/js";
import typescriptEslint from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";
import importAccess from "eslint-plugin-import-access/flat-config";
import simpleImportSort from "eslint-plugin-simple-import-sort";
import unusedImports from "eslint-plugin-unused-imports";
import path from "node:path";
import { fileURLToPath } from "node:url";
import readableTailwind from "eslint-plugin-readable-tailwind";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
  recommendedConfig: js.configs.recommended,
  allConfig: js.configs.all,
});

export default [
  {
    ignores: [
      "coverage",
      ".next",
      "*.config.mjs",
      "tailwind.config.ts",
      "components/ui/**/*",
    ],
  },
  ...fixupConfigRules(
    compat.extends(
      "plugin:@typescript-eslint/recommended",
      "next/core-web-vitals",
      "plugin:import/recommended",
      "plugin:import/warnings",
      "plugin:tailwindcss/recommended",
    ),
  ),
  {
    plugins: {
      "@typescript-eslint": fixupPluginRules(typescriptEslint),
      "simple-import-sort": simpleImportSort,
      "unused-imports": unusedImports,
      "import-access": importAccess,
      "readable-tailwind": readableTailwind,
    },
    languageOptions: {
      parser: tsParser,
      ecmaVersion: "latest",
      sourceType: "module",
      parserOptions: {
        project: "./tsconfig.json",
        tsconfigRootDir: __dirname,
      },
    },
    settings: {
      tailwindcss: {
        callees: ["cn", "cva"], // `cn()` や `cva()` の中のclass名もチェック対象に
      },
    },
    rules: {
      "@typescript-eslint/naming-convention": [
        "error",
        {
          selector: "variable",
          types: ["array", "boolean", "number", "string"],
          format: ["strictCamelCase", "UPPER_CASE"],
        },
        {
          selector: "variable",
          types: ["function"],
          format: ["strictCamelCase", "StrictPascalCase"],
        },
      ],
      "simple-import-sort/imports": "error",
      "simple-import-sort/exports": "error",
      "import/first": "error",
      "import/newline-after-import": "error",
      "import/no-duplicates": "error",
      "@typescript-eslint/consistent-type-exports": "error",
      "import/group-exports": "error",
      "unused-imports/no-unused-imports": "error",
      "import-access/jsdoc": ["error"],
      "no-restricted-imports": [
        "error",
        {
          paths: [
            "sonner",
            "next/link",
            "react-icons",
            "lucide-react",
            "zod",
            { name: "@/components/ui/Form", importNames: ["Form"] },
            {
              name: "@next/third-parties/google",
              importNames: ["sendGAEvent"],
            },
          ],
          patterns: ["react-icons/*"],
        },
      ],
      "tailwindcss/no-custom-classname": [
        "error",
        {
          classRegex:
            "^(class(Name)?|textClassName|iconClassName|innerClassName)$",
          whitelist: ["^[A-Z].*"],
        },
      ],
      "readable-tailwind/multiline": [
        "warn",
        {
          group: "newLine",
        },
      ],
      "no-restricted-syntax": [
        "error",
        {
          selector:
            "CallExpression[callee.object.name='Object'][callee.property.name='keys']",
          message:
            "Do not use Object.keys. Check src/utils/object.ts or add a new utility function.",
        },
      ],
      "@typescript-eslint/no-unused-vars": "off",
      "@typescript-eslint/no-unnecessary-type-assertion": "error",
    },
  },
  {
    files: [
      "src/**/*.stories.tsx",
      "src/**/*Type.ts",
      "src/types/**",
      "src/features/**/*Repository.ts",
      "src/features/**/*Converter.ts",
      "src/features/**/*Constants.ts",
    ],
    rules: {
      "import/group-exports": "off",
    },
  },
  {
    files: ["components/icons/**/*.{ts,tsx}"],
    rules: {
      "no-restricted-imports": "off",
    },
  },
];