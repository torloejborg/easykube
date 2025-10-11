'use strict'

const fsp = require('node:fs/promises')

module.exports.register = function () {
  this.once('contentAggregated', async ({ contentAggregate }) => {
    for (const componentVersionBucket of contentAggregate) {
      componentVersionBucket.files.push(await staticFile('search'))
      componentVersionBucket.files.push(await staticFile('spring-projects'))
    }
  })
}

async function staticFile (name) {
  const resolvedFile = require.resolve(`@springio/antora-extensions/static-pages/${name}`)
  return {
    path: `modules/ROOT/pages/${name}.adoc`,
    contents: Buffer.from(await fsp.readFile(resolvedFile, 'utf8')),
    src: {
      path: `modules/ROOT/pages/${name}.adoc`,
      basename: `${name}.adoc`,
      stem: name,
      extname: '.adoc',
      abspath: resolvedFile,
    },
  }
}
