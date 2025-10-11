'use strict'

const { name: packageName } = require('#package')

module.exports.register = function ({ playbook, config }) {
  this.require(`${packageName}/override-navigation-builder-extension`).register.call(this, { playbook, config })
  this.require(`${packageName}/set-algolia-env-extension`).register.call(this, { playbook, config })
  this.require(`${packageName}/partial-build-extension`).register.call(this, { playbook, config })
  this.require('@antora/atlas-extension').register.call(this, { config })
  this.require(`${packageName}/latest-version-extension`).register.call(this, { config })
  this.require(`${packageName}/inject-collector-cache-config-extension`).register.call(this, { playbook, config })
  this.require('@antora/collector-extension').register.call(this, { config })
  this.require(`${packageName}/root-component-extension`).register.call(this, { config })
  this.require(`${packageName}/root-attachments-extension`).register.call(this, { config })
  this.require(`${packageName}/static-page-extension`).register.call(this, { config })
}
