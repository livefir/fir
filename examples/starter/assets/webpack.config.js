const path = require('path')
const glob = require('glob-all')
const MiniCssExtractPlugin = require('mini-css-extract-plugin')
const CssMinimizerPlugin = require("css-minimizer-webpack-plugin");
const PurgeCssPlugin = require('purgecss-webpack-plugin')
const CopyPlugin = require('copy-webpack-plugin');
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin;
const Dotenv = require('dotenv-webpack');

const PATHS = {
    html: path.join(__dirname, '../templates'),
    src: path.join(__dirname, 'src')
}

const env = process.env.NODE_ENV;


module.exports = {
    mode: env === 'production' || env === 'none' ? env : 'development',
    devtool: "source-map",
    entry: './src/index.js',
    output: {
        path: path.resolve(__dirname, '../public'),
        filename: 'assets/js/bundle.js'
    },
    module: {
        rules: [
            {
                test: /\.js$/,
                exclude: [
                    /node_modules/
                ],
                use: [
                    {loader: "babel-loader"}
                ]
            },
            {
                test: /\.css$/,
                use: [
                    /**
                     * MiniCssExtractPlugin doesn't support HMR.
                     * For developing, use 'style-loader' instead.
                     * */
                    env === 'production' ? MiniCssExtractPlugin.loader : 'style-loader',
                    'css-loader',
                ]
            },
            {
                test: /\.scss$/,
                use: [
                    MiniCssExtractPlugin.loader,
                    {
                        loader: 'css-loader'
                    },
                    {
                        loader: 'sass-loader',
                        options: {
                            // options...
                        }
                    }
                ]
            }]
    },
    plugins: [
        new MiniCssExtractPlugin({
            filename: 'assets/css/styles.css'
        }),
        new CopyPlugin({
            patterns: [
                { from: 'images', to: 'assets/images' },
            ],
        }),

        // new BundleAnalyzerPlugin(),
    ]
};

if (env !== "production") {
    module.exports.plugins.push(new Dotenv({
        path: '.env.dev'
    }))
}

if (env === 'production') {
    module.exports.plugins.push(
        new CssMinimizerPlugin()
    );

    module.exports.plugins.push(
        new PurgeCssPlugin({
            paths: glob.sync([
                    `${PATHS.html}/**/*`,
                    `${PATHS.src}/**/*`],
                {nodir: true}),
        }),
    );

    module.exports.plugins.push( new Dotenv())
}