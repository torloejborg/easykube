'use strict'

const fs = require('node:fs')
const { promises: fsp } = fs
const ospath = require('node:path')
const { XMLParser } = require('fast-xml-parser')

const ExtractRepositoryPathRx = /^(?:\w+@.+?:|\w+:\/\/.+?\/)(.+?)(?:\.git)?$/

module.exports.register = function ({ config = {} }) {
  const refname = config.refname || process.env.BUILD_REFNAME
  if (!refname) return
  this.once('playbookBuilt', async ({ playbook }) => {
    const { concat: get } = this.require('simple-get')
    const asciidocAttrs = ((playbook.asciidoc ||= {}).attributes ||= {})
    const siteManifestUrl = asciidocAttrs['primary-site-manifest-url'] || `${playbook.site.url}/site-manifest.json`
    const siteManifestData = await (siteManifestUrl.startsWith('https://') || siteManifestUrl.startsWith('http://')
      ? download(get, siteManifestUrl)
      : fsp.readFile(siteManifestUrl)
    ).then(JSON.parse)
    const isAuthorMode = refname === 'HEAD'
    const isBranch = refname === 'main' || refname.endsWith('.x') || isAuthorMode
    let version = config.version || process.env.BUILD_VERSION
    if (!version) {
      if (isAuthorMode) {
        throw new Error('When using author mode version is required. Specify config.version or env BUILD_VERSION')
      }
      const { rawgitUrl = 'https://raw.githubusercontent.com' } = config
      const repositoryPath = await getRepositoryPath(this.require('isomorphic-git'), playbook.dir)
      const baseUrl = `${rawgitUrl}/${repositoryPath}/${refname}`
      version = await getMavenVersion(get, baseUrl, refname).catch((me) =>
        getGradleVersion(get, baseUrl, refname).catch((ge) => {
          throw new Error(
            `Could not obtain the version from a pom.xml (Error: ${me}) or gradle.properties (Error: ${ge})`,
            { cause: { maven: me, gradle: ge } }
          )
        })
      )
    }
    if (isBranch && version.endsWith('-SNAPSHOT')) version = version.slice(0, -9)
    const versionsInManifest = Object.values(siteManifestData.components)[0].versions
    if (!isAuthorMode && !(version in versionsInManifest && isBranch === !!versionsInManifest[version].prerelease)) {
      const category = require('path').basename(module.id, '.js')
      await fsp.writeFile(ospath.join(playbook.dir, '.full-build'), '')
      console.log(`${category}: version ${version} not previously built; reverting to full build`)
      return
    }
    const sourcesLength = playbook.content.sources.length
    if (sourcesLength === 1) {
      Object.assign(
        playbook.content.sources[0],
        isBranch
          ? {
              branches: [refname],
              tags: [],
            }
          : {
              branches: [],
              tags: [refname],
            }
      )
      if (isAuthorMode) {
        const authorUrl = resolveAuthorUrl(playbook.dir)
        Object.assign(playbook.content.sources[0], { url: authorUrl })
      }
    } else {
      console.log(`sources.length = ${sourcesLength} so deferring filtering till contentAggregated`)
    }
    Object.assign(asciidocAttrs, { 'primary-site-url': '.', 'primary-site-manifest-url': siteManifestUrl })
    this.updateVariables({ playbook })
  })
  this.once('contentAggregated', async ({ playbook, contentAggregate }) => {
    if (contentAggregate.length <= 1) {
      return
    }
    if (fs.existsSync(ospath.join(playbook.dir, '.full-build'), '')) {
      return
    }
    const indexToSave = contentAggregate.findIndex((c) => c.origins[0].branch === refname)
    if (indexToSave < 0) {
      throw new Error(`Could not find ${refname}`)
    }
    const toSave = contentAggregate[indexToSave]
    contentAggregate.splice(0, contentAggregate.length)
    contentAggregate.push(toSave)
  })
}

function download (get, url) {
  return new Promise((resolve, reject) =>
    get({ url }, (err, response, contents) => {
      if (err) return reject(err)
      if (response.statusCode === 200) return resolve(contents)
      const message = `${url} returned response code ${response.statusCode} (${response.statusMessage})`
      reject(Object.assign(new Error(message), { name: 'HTTPError' }))
    })
  )
}

function getGradleVersion (get, baseUrl, refname) {
  const gradlePropertiesUrl = `${baseUrl}/gradle.properties`
  return download(get, gradlePropertiesUrl).then((contents) => extractGradleVersion(contents, refname))
}

function extractGradleVersion (contents, fallback) {
  for (const line of (contents.toString().trimEnd() + `\nversion=${fallback}`).split('\n')) {
    if (line.startsWith('version=')) return line.slice(8)
  }
}

function getMavenVersion (get, baseUrl, refname) {
  const pomUrl = `${baseUrl}/pom.xml`
  return download(get, pomUrl).then((contents) => extractMavenVersion(contents, refname))
}

function extractMavenVersion (contents, fallback) {
  const parser = new XMLParser()
  const jObj = parser.parse(contents)
  return jObj.project.version
}

function getRepositoryPath (git, dir) {
  return resolveGitdir(git, dir).then((gitdir) =>
    git
      .listRemotes({ fs, gitdir })
      .then((remotes) => ExtractRepositoryPathRx.exec(remotes.find(({ remote }) => remote === 'origin').url)[1])
  )
}

function resolveGitdir (git, dir, dotgit = ospath.join(dir, '.git')) {
  return fsp
    .stat(dotgit)
    .then((stat) => {
      if (stat.isDirectory()) return dotgit
      return fsp
        .readFile(dotgit, 'utf8')
        .then((contents) => contents.substr(8).trimRight())
        .then((worktreeGitdir) =>
          fsp
            .readFile(ospath.join(worktreeGitdir, 'commondir'), 'utf-8')
            .then((commondir) => ospath.join(worktreeGitdir, commondir.trimRight()))
        )
    })
    .catch(() => resolveGitdir(git, ospath.dirname(dir)))
}

function resolveAuthorUrl (dir) {
  const gitRoot = resolveGitRoot(dir)
  if (gitRoot === dir) {
    return '.'
  }
  const relpath = ospath.relative(gitRoot, dir)
  const segments = [...relpath].filter((c) => c === ospath.sep).length + 1
  return '.' + '/..'.repeat(segments)
}

function resolveGitRoot (dir, dotgit = ospath.join(dir, '.git')) {
  return fs.existsSync(dotgit) ? dir : resolveGitRoot(ospath.dirname(dir))
}
