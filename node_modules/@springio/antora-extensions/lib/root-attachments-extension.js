'use strict'

module.exports.register = function ({ config = {} }) {
  this.once('contentClassified', ({ contentCatalog }) => {
    contentCatalog.findBy({ family: 'attachment' }).forEach((file) => {
      removeRootComponentNameFromFile(file)
    })
  })
}

function removeRootComponentNameFromFile (file) {
  if (file.out) {
    file.out.rootPath = removeFirstSegment(file.out.rootPath)
    file.out.moduleRootPath = removeFirstSegment(file.out.moduleRootPath)
    file.out.dirname = removeAttachmentsFromPath(file.out.dirname)
    file.out.path = removeAttachmentsFromPath(file.out.path)
  }
  if (file.pub) {
    if (file.pub.rootPath) {
      file.pub.rootPath = removeFirstSegment(file.pub.rootPath)
    }
    if (file.pub.moduleRootPath) {
      file.pub.moduleRootPath = removeFirstSegment(file.pub.moduleRootPath)
    }
    file.pub.url = removeAttachmentsFromPath(file.pub.url)
  }
}

function removeFirstSegment (path) {
  return path ? path.split('/').slice(1).join('/') || '.' : path
}

function removeAttachmentsFromPath (path) {
  return path.replace('/_attachments', '').replace('_attachments/', '')
}
