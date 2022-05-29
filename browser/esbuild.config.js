import esbuildServe from "esbuild-serve";

esbuildServe(
  {
    logLevel: "info",
    entryPoints: ["index.js"],
    bundle: true,
    outfile: "dist/cdn.js",
    platform: "browser",
    define: { CDN: true }
  },
  { root: "dist", port: 8000 }
);