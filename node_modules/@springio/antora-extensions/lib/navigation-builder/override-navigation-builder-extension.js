'use strict'

const buildNavigation = require('./build-navigation')

module.exports.register = function () {
  this.replaceFunctions({
    buildNavigation (contentCatalog, siteAsciiDocConfig) {
      return buildNavigation.call(this, contentCatalog, siteAsciiDocConfig)
    },
  })
}
