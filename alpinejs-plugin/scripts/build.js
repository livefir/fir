//copied from https://github.com/victoryoalli/alpinejs-plugin-blueprint/blob/main/scripts/build.js
//CDN
build({
    entryPoints: [`builds/cdn.js`],
    outfile: `dist/fir.min.js`,
    bundle: true,
    minify: true,
    sourcemap: false,
    platform: 'browser',
    define: { CDN: true },
})
//Module
build({
    entryPoints: [`builds/module.js`],
    outfile: `dist/fir.esm.js`,
    bundle: true,
    bundle: true,
    platform: 'neutral',
    mainFields: ['main', 'module'],
})
build({
    entryPoints: [`builds/module.js`],
    outfile: `dist/fir.cjs.js`,
    bundle: true,
    target: ['node10.4'],
    platform: 'node',
})


function build(options) {
    options.define || (options.define = {})
    options.define['process.env.NODE_ENV'] = process.argv.includes('--watch') ? `'production'` : `'development'`

    return require('esbuild').build({
        watch: process.argv.includes('--watch'),
        // external: ['alpinejs'],
        ...options,
    }).catch(() => process.exit(1))
}