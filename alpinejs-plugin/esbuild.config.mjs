import esbuildServe from 'esbuild-serve'

esbuildServe(
    {
        logLevel: 'info',
        entryPoints: ['builds/cdn.js'],
        bundle: true,
        outfile: 'dist/cdn.js',
        platform: 'browser',
        minify: false,
        define: { CDN: true },
    },
    { root: 'dist', port: 8000 }
)
