import { defineConfig } from "orval";

export default defineConfig({
  stepOfficialWebsite: {
    input: "../backend/docs/v3/openapi.yaml",
    output: {
      target: "./src/api/__generated__/",
      schemas: "./src/api/__generated__/schemas",
      client: "react-query",
      mode: "tags-split",
      override: {
        mutator: {
          path: "./src/api/customAxios.ts",
          name: "customAxios",
          default: true,
        },
        query: {
          useQuery: true,
          usePrefetch: true,
        },
      },
    },
  },
});
