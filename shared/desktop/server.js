// @flow
// Builds our code, serves changes if NO_SERVER is false
const path = require('path')
const webpack = require('webpack')
const WebpackDevServer = require('webpack-dev-server')
const getenv = require('getenv')

const config = getenv.boolish('DUMB', false)
  ? Object.assign({}, require('./webpack.config.dumb'))
  : Object.assign({}, require('./webpack.config.development'))
const PORT = 4000

const compiler = webpack(config)

// Just build output files and don't run a hot server
const NO_SERVER = getenv.boolish('NO_SERVER', false)
const KEYBASE_VERBOSE_WEBPACK = getenv.boolish('KEYBASE_VERBOSE_WEBPACK', false)

if (NO_SERVER) {
  console.log('Starting local file build')
  compiler.run(function(err, stats) {
    if (err) {
      throw err
    }
    var jsonStats = stats.toJson()
    if (jsonStats.errors.length > 0) {
      throw jsonStats.errors.join('\n')
    }

    console.log(stats)

    if (jsonStats.warnings.length > 0) {
      console.log('Warnings: ', jsonStats.warnings.join('\n'))
    }
  })
} else {
  const server = new WebpackDevServer(compiler, {
    compress: true,
    contentBase: path.join(__dirname, 'dist'),
    hot: true,
    lazy: false,
    publicPath: 'http://localhost:4000/dist/',
    quiet: false,
    stats: 'verbose',
    // {
    // chunkModules: KEYBASE_VERBOSE_WEBPACK,
    // colors: true,
    // },
  })

  server.listen(PORT, 'localhost', err => {
    if (err) {
      console.log(err)
      return
    }

    console.log(`Listening at http://localhost:${PORT}`)
  })
}
