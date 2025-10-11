'use strict'

const fsp = require('node:fs/promises')
const ospath = require('node:path')
const resolvedConfigPath = require.resolve('@springio/antora-extensions/docsearch-config.json.hbs')

/**
 * An Antora extension that generates the docsearch config file from a Handlebars template and publishes it with the
 * site, where the scraper job can retrieve it.
 */
module.exports.register = function ({ config: { templatePath, indexName, rootComponentName } }) {
  const expandPath = this.require('@antora/expand-path-helper')
  const handlebars = this.require('handlebars').create()
  handlebars.registerHelper('eq', (a, b) => a === b)
  handlebars.registerHelper('and', (a, b) => a && b)
  handlebars.registerHelper('defined', (a, message) => {
    if (a) {
      return a
    }
    throw new Error(message)
  })

  this.on('beforePublish', async ({ playbook, contentCatalog, siteCatalog }) => {
    templatePath = templatePath ? expandPath(templatePath, { dot: playbook.dir }) : resolvedConfigPath
    const templateSrc = await fsp.readFile(templatePath, 'utf8')
    const templateBasename = ospath.basename(templatePath)
    const template = handlebars.compile(templateSrc, { noEscape: true, preventIndent: true, srcName: templateBasename })
    const latestVersions = contentCatalog.getComponentsSortedBy('name').reduce((accum, component) => {
      component.versions.forEach((version) => version.versionSegment !== undefined && accum.push(version))
      return accum
    }, [])
    const stopPages = contentCatalog.getPages((page) => {
      return page.out && ('page-archived' in page.asciidoc.attributes || 'page-noindex' in page.asciidoc.attributes)
    }, [])
    if (!latestVersions.length) {
      throw new Error('The content catalog does not contain any versions to index.')
    }
    const defaultedIndexName = indexName || `${latestVersions[0].name}-docs`
    const compiled = template({
      indexName: defaultedIndexName,
      rootComponentName,
      latestVersions,
      site: playbook.site,
      stopPages,
    })
    siteCatalog.addFile({ contents: Buffer.from(compiled), out: { path: 'docsearch-config.json' } })
  })
}
