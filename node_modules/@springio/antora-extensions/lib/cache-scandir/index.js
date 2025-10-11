'use strict'

const copyRecursive = require('./copy-recursive')

const [, , ...args] = process.argv
const [scanDir, cacheDir] = args

copyRecursive(scanDir, cacheDir)
