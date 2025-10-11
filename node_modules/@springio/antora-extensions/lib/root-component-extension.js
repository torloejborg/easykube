'use strict'

module.exports.register = function ({ config = {} }) {
  this.once('contentClassified', ({ contentCatalog }) => {
    const rootComponentName = getRootComponentName(config)
    contentCatalog.findBy({ component: rootComponentName }).forEach((file) => {
      removeRootComponentNameFromFile(rootComponentName, file)
    })
    const rootComponent = contentCatalog.getComponent(rootComponentName)
    rootComponent?.versions?.forEach((version) => {
      version.url = removeRootComponentNameFromUrl(rootComponentName, version.url)
    })
  })
  this.once('documentsConverted', ({ contentCatalog }) => {
    const rootComponentName = getRootComponentName(config)
    contentCatalog.findBy({ component: rootComponentName, family: 'alias' }).forEach((file) => {
      removeRootComponentNameFromFile(rootComponentName, file)
    })
  })
}

function getRootComponentName (config) {
  const rootComponentName = config.rootComponentName
  if (!rootComponentName) {
    throw new Error('Missing required configuration attribute root_component_name for root-component-extension')
  }
  return rootComponentName
}

function removeRootComponentNameFromFile (rootComponentName, file) {
  if (file.out) {
    file.out.rootPath = fixRootPathForRootComponentNameAndUrl(file.out.rootPath, rootComponentName, file.out.path)
    file.out.dirname = removeRootComponentNameFromUrl(rootComponentName, file.out.dirname)
    file.out.path = removeRootComponentNameFromUrl(rootComponentName, file.out.path)
  }
  if (file.pub) {
    if (file.pub.rootPath) {
      file.pub.rootPath = fixRootPathForRootComponentNameAndUrl(file.pub.rootPath, rootComponentName, file.pub.url)
    }
    file.pub.url = removeRootComponentNameFromUrl(rootComponentName, file.pub.url)
  }
  if (file.rel) {
    if (file.rel.pub) {
      file.rel.pub.rootPath = fixRootPathForRootComponentNameAndUrl(
        file.rel.pub.rootPath,
        rootComponentName,
        file.rel.pub.url
      )
      file.rel.pub.url = removeRootComponentNameFromUrl(rootComponentName, file.rel.pub.url) || '/'
    }
  }
}

function urlStartsWithRootComponentName (url, rootComponentName) {
  return (
    url.startsWith(`/${rootComponentName}/`) ||
    url.startsWith(`${rootComponentName}/`) ||
    url === `/${rootComponentName}`
  )
}

function removeRootComponentNameFromUrl (rootComponentName, url) {
  if (urlStartsWithRootComponentName(url, rootComponentName)) {
    return url.slice(rootComponentName.length + 1)
  }
  return url
}

function fixRootPathForRootComponentNameAndUrl (rootPath, rootComopnentName, url) {
  if (urlStartsWithRootComponentName(url, rootComopnentName)) {
    return rootPath.split('/').slice(1).join('/') || '.'
  }
  return rootPath
}
