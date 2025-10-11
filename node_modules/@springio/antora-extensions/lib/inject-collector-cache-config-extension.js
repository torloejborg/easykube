'use strict'

const expandPath = require('@antora/expand-path-helper')
const ospath = require('path')
const resolvedCopyRecursiveJs = require.resolve('@springio/antora-extensions/cache-scandir')
const { createHash } = require('crypto')
const archiver = require('archiver')

module.exports.register = function ({ playbook, config = {} }) {
  const logger = this.getLogger('inject-collector-cache-config-extension')
  const siteUrl = playbook.site?.url
  const outputDir = ospath.join(playbook.dir, 'build/antora/inject-collector-cache-config-extension/.cache')
  const configuredBaseCacheUrl = config.baseCacheUrl
  if (!siteUrl && !configuredBaseCacheUrl) {
    throw new Error(
      "One of playbook site.url or inject-collector-cache-config-extension plugin's base_cache_url property is required."
    )
  }
  const baseCacheUrl = configuredBaseCacheUrl || siteUrl + '/.cache'
  const getUserCacheDir = this.require('cache-directory')
  const expandPath = this.require('@antora/expand-path-helper')
  const fs = this.require('fs')
  const decompress = this.require('decompress')
  const { concat: get } = this.require('simple-get')
  if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true })
  }
  const zipInfo = []
  this.once('contentAggregated', async ({ playbook, contentAggregate }) => {
    for (const { origins } of contentAggregate) {
      for (const origin of origins) {
        const { url, gitdir, refhash, worktree, startPath = '.', descriptor, tag, refname } = origin
        const baseCacheDir = getBaseCacheDir(getUserCacheDir, playbook)
        const repositoryCachDirName = generateWorktreeFolderName({ url, gitdir, worktree })
        const baseCollectorCacheCacheDir = ospath.join(baseCacheDir, 'collector-cache', repositoryCachDirName)
        const collectorCacheDir = ospath.join(baseCacheDir, 'collector')
        const worktreeDir = worktree || ospath.join(collectorCacheDir, repositoryCachDirName)
        const expandPathContext = { base: worktreeDir, cwd: worktreeDir, dot: ospath.join(worktreeDir, startPath) }
        const collectorConfig = descriptor?.ext?.collector
        if (collectorConfig !== undefined && tag !== undefined) {
          const shortref = refhash.slice(0, 7)
          const cacheDirName = `${shortref}-${refname}`
          const cacheDir = ospath.join(baseCollectorCacheCacheDir, cacheDirName)
          const zipFileName = `${cacheDirName}.zip`
          const zipCacheFile = ospath.join(outputDir, zipFileName)
          if (!fs.existsSync(baseCollectorCacheCacheDir)) {
            fs.mkdirSync(baseCollectorCacheCacheDir, { recursive: true })
          }
          if (!fs.existsSync(cacheDir)) {
            // try and restore from URL by downloading zip
            const cacheUrl = `${baseCacheUrl}/${cacheDirName}.zip`
            const content = await download(get, cacheUrl).then((content) => content)
            if (content) {
              fs.writeFileSync(zipCacheFile, content)
              await decompress(zipCacheFile, ospath.join(baseCollectorCacheCacheDir, cacheDirName)).then((files) =>
                logger.debug(`Successfully unzipped ${zipCacheFile}.`)
              )
              logger.info(`Successfully restored cache from ${cacheUrl}`)
            } else {
              logger.info(`Unable to restore cache from ${cacheUrl}`)
            }
          }
          if (fs.existsSync(cacheDir)) {
            // use the cache
            origin.descriptor.ext.collector = {
              scan: {
                dir: cacheDir,
              },
            }
            logger.info(`Use the cache found at ${cacheDir}`)
          } else {
            const cachedConfig = []
            const normalizedCollectorConfig = Array.isArray(collectorConfig) ? collectorConfig : [collectorConfig]
            origin.descriptor.ext.collector = cachedConfig
            normalizedCollectorConfig.forEach((collector) => {
              const { scan: scanConfig = [] } = collector
              // cache the output of the build
              const scanDir = expandPath(scanConfig.dir, expandPathContext)
              logger.info(`Configuring collector to cache '${scanDir}' at '${cacheDir}'`)
              const cachedCollectorConfig = createCachedCollectorConfig(scanDir, cacheDir)
              cachedConfig.push(collector)
              cachedConfig.push.apply(cachedConfig, cachedCollectorConfig)
            })
            // add the zip of cache to be published
            zipInfo.push({ cacheDir, zipCacheFile })
          }
        }
      }
    }
  })
  this.once('beforePublish', async () => {
    for (const info of zipInfo) {
      await zip(fs, logger, info.cacheDir, info.zipCacheFile)
    }
  })
}

function download (get, url) {
  return new Promise((resolve, reject) =>
    get({ url }, (err, response, contents) => {
      if (response?.statusCode === 404) return resolve(undefined)
      if (err) return reject(err)
      if (response?.statusCode === 200) return resolve(contents)
      const message = `${url} returned response code ${response?.statusCode} (${response?.statusMessage})`
      reject(Object.assign(new Error(message), { name: 'HTTPError' }))
    })
  )
}

function getBaseCacheDir (getUserCacheDir, { dir: dot, runtime: { cacheDir: preferredDir } = {} }) {
  return preferredDir == null
    ? getUserCacheDir(`antora${process.env.NODE_ENV === 'test' ? '-test' : ''}`) || ospath.join(dot, '.cache/antora')
    : expandPath(preferredDir, { dot })
}

function generateWorktreeFolderName ({ url, gitdir, worktree }) {
  if (worktree === undefined) return ospath.basename(gitdir, '.git')
  return `${url.substr(url.lastIndexOf('/') + 1)}-${createHash('sha1').update(url).digest('hex')}`
}

function createCachedCollectorConfig (scanDir, cacheDir) {
  return [
    {
      run: {
        command: `node '${resolvedCopyRecursiveJs}' '${scanDir}' '${cacheDir}'`,
      },
    },
  ]
}

const zip = async function (fs, logger, src, destination) {
  const path = require('path')
  const destParent = path.dirname(destination)
  if (!fs.existsSync(destParent)) {
    fs.mkdirs(destParent, { recursive: true })
  }
  const output = fs.createWriteStream(destination)
  const archive = archiver('zip', {
    zlib: { level: 9 }, // Sets the compression level.
  })
  // listen for all archive data to be written
  // 'close' event is fired only when a file descriptor is involved
  output.on('close', function () {
    logger.info(archive.pointer() + ' total bytes')
    logger.info('archiver has been finalized and the output file descriptor has closed.')
  })

  // This event is fired when the data source is drained no matter what was the data source.
  // It is not part of this library but rather from the NodeJS Stream API.
  // @see: https://nodejs.org/api/stream.html#stream_event_end
  output.on('end', function () {
    logger.info('Data has been drained')
  })

  // good practice to catch warnings (ie stat failures and other non-blocking errors)
  archive.on('warning', function (err) {
    if (err.code === 'ENOENT') {
      // log warning
    } else {
      // throw error
      throw err
    }
  })

  // good practice to catch this error explicitly
  archive.on('error', function (err) {
    throw err
  })

  // pipe archive data to the file
  archive.pipe(output)

  archive.directory(src, false)

  await archive.finalize()

  logger.info(`Saving ${src} into ${destination}`)
}
