import js from "@eslint/js";
import globals from "globals";
import reactHooks from "eslint-plugin-react-hooks";
import reactRefresh from "eslint-plugin-react-refresh";
import tseslint from "typescript-eslint";

export default tseslint.config(
  { ignores: ["dist"] },
  {
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    files: ["**/*.{ts,tsx}"],
    languageOptions: {
      ecmaVersion: 2022,
      globals: globals.browser,
    },
    plugins: {
      "react-hooks": reactHooks,
      "react-refresh": reactRefresh,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      "react-refresh/only-export-components": ["warn", { allowConstantExport: true }],

      // Physical properties break RTL. Retrofitting logical ones later is
      // miserable, so they are refused from the first commit.
      "no-restricted-syntax": [
        "error",
        {
          selector:
            "Literal[value=/(^|\\s)(ml|mr|pl|pr|border-l|border-r|rounded-l|rounded-r|left|right|text-left|text-right)-/]",
          message:
            "Use logical properties: ms/me, ps/pe, border-s/border-e, start/end, text-start/text-end.",
        },
      ],
    },
  },
);
