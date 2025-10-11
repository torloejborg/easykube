'use strict'

const { name: packageName } = require('../package.json')
const fs = require('fs')
const crypto = require('crypto')
const { promises: fsp } = fs
const LazyReadable = require('./lazy-readable')
const MultiFileReadStream = require('./multi-file-read-stream')
const ospath = require('path')
const template = require('./template')

function register ({ config: { rows, cols, autoPlay, ...unknownOptions } }) {
  const logger = this.getLogger(packageName)

  if (Object.keys(unknownOptions).length) {
    const keys = Object.keys(unknownOptions)
    throw new Error(`Unrecognized option${keys.length > 1 ? 's' : ''} specified for ${packageName}: ${keys.join(', ')}`)
  }

  const defaultOptions = { rows, cols, autoPlay }

  this.on('uiLoaded', async ({ playbook, uiCatalog }) => {
    playbook.env.SITE_ASCIINEMA_PROVIDER = 'asciinema'
    const asciinemaDir = 'asciinema'
    const uiOutputDir = playbook.ui.outputDir
    vendorJsFile(
      uiCatalog,
      logger,
      uiOutputDir,
      'asciinema-player/dist/bundle/asciinema-player.min.js',
      'asciinema-player.js'
    )
    vendorCssFile(
      uiCatalog,
      logger,
      uiOutputDir,
      'asciinema-player/dist/bundle/asciinema-player.css',
      'asciinema-player.css'
    )

    const asciinemaLoadScriptsPartialPath = 'asciinema-load.hbs'
    if (!uiCatalog.findByType('partial').some(({ path }) => path === asciinemaLoadScriptsPartialPath)) {
      const asciinemaLoadScriptsPartialFilepath = ospath.join(__dirname, asciinemaDir, asciinemaLoadScriptsPartialPath)
      uiCatalog.addFile({
        contents: Buffer.from(template(await fsp.readFile(asciinemaLoadScriptsPartialFilepath, 'utf8'), {})),
        path: asciinemaLoadScriptsPartialPath,
        stem: 'asciinema-load-scripts',
        type: 'partial',
      })
    }

    const asciinemaCreateScriptsPartialPath = 'asciinema-create.hbs'
    if (!uiCatalog.findByType('partial').some(({ path }) => path === asciinemaCreateScriptsPartialPath)) {
      const asciinemaCreateScriptsPartialFilepath = ospath.join(
        __dirname,
        asciinemaDir,
        asciinemaCreateScriptsPartialPath
      )
      uiCatalog.addFile({
        contents: Buffer.from(template(await fsp.readFile(asciinemaCreateScriptsPartialFilepath, 'utf8'), {})),
        path: asciinemaCreateScriptsPartialPath,
        stem: 'asciinema-create-scripts',
        type: 'partial',
      })
    }

    const asciinemaStylesPartialPath = 'asciinema-styles.hbs'
    if (!uiCatalog.findByType('partial').some(({ path }) => path === asciinemaStylesPartialPath)) {
      const asciinemaStylesPartialFilepath = ospath.join(__dirname, asciinemaDir, asciinemaStylesPartialPath)
      uiCatalog.addFile({
        contents: Buffer.from(template(await fsp.readFile(asciinemaStylesPartialFilepath, 'utf8'), {})),
        path: asciinemaStylesPartialPath,
        stem: 'asciinema-styles',
        type: 'partial',
      })
    }

    const splitHelperPartialPath = 'asciinema-split-helper.js'
    const splitHelperPartialFilepath = ospath.join(__dirname, asciinemaDir, splitHelperPartialPath)
    uiCatalog.addFile({
      contents: Buffer.from(template(await fsp.readFile(splitHelperPartialFilepath, 'utf8'), {})),
      path: 'helpers/' + splitHelperPartialPath,
      stem: 'asciinema-split',
      type: 'helper',
    })

    const optionsHelperPartialPath = 'asciinema-options-helper.js'
    const optionsHelperPartialFilepath = ospath.join(__dirname, asciinemaDir, optionsHelperPartialPath)
    uiCatalog.addFile({
      contents: Buffer.from(template(await fsp.readFile(optionsHelperPartialFilepath, 'utf8'), {})),
      path: 'helpers/' + optionsHelperPartialPath,
      stem: 'asciinema-options',
      type: 'helper',
    })
  })

  this.on('contentClassified', async ({ siteAsciiDocConfig, uiCatalog }) => {
    if (!siteAsciiDocConfig.extensions) siteAsciiDocConfig.extensions = []
    siteAsciiDocConfig.extensions.push({
      register: (registry, _context) => {
        registry.block('asciinema', processAsciinemaBlock(uiCatalog, defaultOptions, _context))
        return registry
      },
    })
  })
}

function processAsciinemaBlock (uiCatalog, defaultOptions, context) {
  return function () {
    this.onContext(['listing', 'literal'])
    this.positionalAttributes(['target', 'format'])
    this.process((parent, reader, attrs) => {
      const { file } = context
      const source = reader.getLines().join('\n')
      return toBlock(attrs, parent, source, this, uiCatalog, defaultOptions, file)
    })
  }
}

const fromHash = (hash) => {
  const object = {}
  const data = hash.$$smap
  for (const key in data) {
    object[key] = data[key]
  }
  return object
}

const toBlock = (attrs, parent, source, context, uiCatalog, defaultOptions, file) => {
  if (typeof attrs === 'object' && '$$smap' in attrs) {
    attrs = fromHash(attrs)
  }
  const doc = parent.getDocument()
  const subs = attrs.subs
  if (subs) {
    source = doc.$apply_subs(attrs.subs, doc.$resolve_subs(subs))
  }
  const idAttr = attrs.id ? ` id="${attrs.id}"` : ''
  const classAttr = attrs.role ? `${attrs.role} asciinemablock` : 'asciinemablock'

  const block = context.$create_pass_block(parent, '', Opal.hash(attrs))

  const title = attrs.title
  if (title) {
    block.title = title
    delete block.caption
    const caption = attrs.caption
    delete attrs.caption
    block.assignCaption(caption, 'figure')
  }

  const asciinemaId = crypto.createHash('md5').update(source, 'utf8').digest('hex')
  if (file.asciidoc.attributes['page-asciinemacasts']) {
    file.asciidoc.attributes['page-asciinemacasts'] =
      file.asciidoc.attributes['page-asciinemacasts'] + ',' + asciinemaId
  } else {
    file.asciidoc.attributes['page-asciinemacasts'] = asciinemaId
  }

  const castFilePath = '_asciinema/' + asciinemaId + '.cast'
  const castFile = uiCatalog.findByType('asset').some(({ path }) => path === castFilePath)

  // same file is either duplicate in other page or in other branch
  // no need to come up with real global guid
  if (!castFile) {
    uiCatalog.addFile({
      contents: Buffer.from(source),
      path: castFilePath,
      type: 'asset',
      out: { path: castFilePath },
    })
  }

  const asciinemaOptions = JSON.stringify(buildOptions(attrs, defaultOptions))
  file.asciidoc.attributes['page-asciinema-options-' + asciinemaId] = asciinemaOptions

  const titleElement = title ? `<div class="title">${block.caption}${title}</div>` : ''
  const style = `${Object.hasOwn(attrs, 'width') ? `width: ${attrs.width}px;` : ''} ${
    Object.hasOwn(attrs, 'height') ? `height: ${attrs.height}px;` : ''
  }`
  block.lines = [
    `<div${idAttr} class="${classAttr}">`,
    `<div class="content"><div id="${asciinemaId}" style="${style}"></div></div>`,
    `${titleElement}</div>`,
  ]
  return block
}

function buildOptions (attrs, defaultOptions) {
  const options = {}
  const rows = attrs.rows ? attrs.rows : defaultOptions.rows
  if (rows) {
    options.rows = rows
  }
  const cols = attrs.cols ? attrs.cols : defaultOptions.cols
  if (cols) {
    options.cols = cols
  }
  const autoPlay = attrs.autoPlay ? attrs.autoPlay : defaultOptions.autoPlay
  if (autoPlay) {
    options.autoPlay = autoPlay
  }
  return options
}

function assetFile (
  uiCatalog,
  logger,
  uiOutputDir,
  assetDir,
  basename,
  assetPath = assetDir + '/' + basename,
  contents = new LazyReadable(() => fs.createReadStream(ospath.join(__dirname, '../data', assetPath))),
  overwrite = false
) {
  const outputDir = uiOutputDir + '/' + assetDir
  const existingFile = uiCatalog.findByType('asset').some(({ path }) => path === assetPath)
  if (existingFile) {
    if (overwrite) {
      logger.warn(`Please remove the following file from your UI since it is managed by ${packageName}: ${assetPath}`)
      existingFile.contents = contents
      delete existingFile.stat
    } else {
      logger.info(`The following file already exists in your UI: ${assetPath}, skipping`)
    }
  } else {
    uiCatalog.addFile({
      contents,
      type: 'asset',
      path: assetPath,
      out: { dirname: outputDir, path: outputDir + '/' + basename, basename },
    })
  }
}

function vendorJsFile (uiCatalog, logger, uiOutputDir, requireRequest, basename = requireRequest.split('/').pop()) {
  let contents
  if (Array.isArray(requireRequest)) {
    const filepaths = requireRequest.map(require.resolve)
    contents = new LazyReadable(() => new MultiFileReadStream(filepaths))
  } else {
    const filepath = require.resolve(requireRequest)
    contents = new LazyReadable(() => fs.createReadStream(filepath))
  }
  const jsVendorDir = 'js/vendor'
  assetFile(uiCatalog, logger, uiOutputDir, jsVendorDir, basename, jsVendorDir + '/' + basename, contents)
}

function vendorCssFile (uiCatalog, logger, uiOutputDir, requireRequest, basename = requireRequest.split('/').pop()) {
  let contents
  if (Array.isArray(requireRequest)) {
    const filepaths = requireRequest.map(require.resolve)
    contents = new LazyReadable(() => new MultiFileReadStream(filepaths))
  } else {
    const filepath = require.resolve(requireRequest)
    contents = new LazyReadable(() => fs.createReadStream(filepath))
  }
  const jsVendorDir = 'css/vendor'
  assetFile(uiCatalog, logger, uiOutputDir, jsVendorDir, basename, jsVendorDir + '/' + basename, contents)
}

module.exports = { register }
