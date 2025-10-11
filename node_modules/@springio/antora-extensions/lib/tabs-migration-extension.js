'use strict'

const fsp = require('node:fs/promises')

module.exports.register = function ({ config = {} }) {
  this.once('contentClassified', ({ playbook, contentCatalog }) => {
    if (!~playbook.asciidoc.extensions.indexOf('@asciidoctor/tabs')) return
    const { saveResult, unwrapExampleBlock = 'tabs', tabsDelimiterLength = 6, normalize } = config
    const filesToMigrate = contentCatalog
      .getPages(({ out }) => out)
      .concat(contentCatalog.findBy({ family: 'partial' }))
      .filter(
        (candidate) =>
          candidate.mediaType === 'text/asciidoc' &&
          (~candidate.contents.indexOf('role="primary"') ||
            (unwrapExampleBlock === 'always' && ~candidate.contents.indexOf('\n====\n')))
      )
    if (!filesToMigrate.length) return
    const saves = filesToMigrate.reduce((saves, page) => {
      const originalContents = page.contents.toString()
      let lines
      let eof = originalContents.length - originalContents.trimEnd().length
      if (eof) {
        eof = originalContents.slice(originalContents.length - eof)
        lines = originalContents.slice(0, originalContents.length - eof.length).split('\n')
      } else {
        eof = ''
        lines = originalContents.split('\n')
      }
      const blockMetadata = new BlockMetadata()
      const containerStack = new ContainerStack(blockMetadata, '='.repeat(tabsDelimiterLength))
      const contents = lines
        .reduce((accum, line, idx) => {
          let chr0, isTab
          if (containerStack.verbatim) {
            if (containerStack.isTerminus(line)) containerStack.pop()
          } else if (containerStack.isTerminus(line)) {
            if (containerStack.closeBlock(accum) === 'unwrapped') return accum
          } else if (line.length) {
            if (line === '----' || line === '-----' || line === '....') {
              containerStack.push(line, accum, { verbatim: true })
              isTab = containerStack.isTab()
            } else if (line === '====' || line === '======') {
              if (containerStack.shouldUnwrapBlock(unwrapExampleBlock, lines, idx)) {
                containerStack.push(line, accum, { unwrap: true })
                return accum
              }
              containerStack.push(line, accum)
            } else if ((chr0 = line.charAt()) === '[' && line.charAt(line.length - 1) === ']') {
              if (blockMetadata.empty && !containerStack.isTerminus(lines[idx - 1])) containerStack.startNewBlock()
              blockMetadata.store(line)
              if (containerStack.isNewTabs()) containerStack.closeTabsBlock(accum)
              return accum
            } else if (chr0 === '.' && line.length > 1 && line.charAt(1) !== ' ' && !line.startsWith('...')) {
              blockMetadata.store(line)
              if (containerStack.isNewTabs()) containerStack.closeTabsBlock(accum)
              return accum
            } else if (containerStack.current && chr0 === '=' && /^=+ \S/.test(line)) {
              containerStack.closeDanglingBlock(accum)
              accum.push(line)
              return accum
            } else {
              isTab = containerStack.isTab()
            }
            if (isTab) {
              containerStack.startTab(accum)
            } else if (containerStack.inTabs) {
              if (containerStack.inTabs === 'expecting-tab') {
                containerStack.closeTabsBlock(accum)
                containerStack.recordNonTab(accum)
              }
            } else {
              containerStack.recordNonTab(accum)
            }
          } else if (normalize && accum[accum.length - 1] === '') {
            return accum
          } else {
            containerStack.startNewBlock(accum)
          }
          accum.push(line)
          return accum
        }, [])
        .join('\n')
        .concat(containerStack.inTabs ? containerStack.closeTabsBlock() : '')
        .concat(blockMetadata.drain())
        .concat(eof)
      if (contents === originalContents) return saves
      page.contents = Buffer.from(contents)
      return saveResult && page.src.abspath ? saves.concat(fsp.writeFile(page.src.abspath, contents, 'utf8')) : saves
    }, [])
    return saves.length ? Promise.all(saves) : undefined
  })
}

class BlockMetadata {
  constructor (entries = []) {
    this.entries = entries
  }

  get attrlist () {
    return this.entries.find((it) => it.type === 'attrlist')?.value
  }

  get empty () {
    return !this.entries.length
  }

  get title () {
    return this.entries.find((it) => it.type === 'title')?.value
  }

  clear () {
    return this.entries.splice(0)
  }

  clone () {
    return new BlockMetadata(this.entries.slice())
  }

  drain (accum) {
    const lines = this.getLines(true)
    if (!accum) return lines.map((l) => '\n' + l)
    accum.push(...lines)
  }

  getLines (clear, entryFilter) {
    let entries = clear ? this.clear() : this.entries
    if (entryFilter) entries = entries.filter(entryFilter)
    return entries.map((it) => it.line)
  }

  hasRole (role) {
    return this.attrlist?.includes(`role="${role}"`)
  }

  hasStyle () {
    return !!this.attrlist?.split(/[%#.,]/, 1)[0]
  }

  isTab (role) {
    return this.title && this.hasRole(role)
  }

  store (line) {
    if (line.charAt() === '.') {
      this.entries.push({ line, type: 'title', value: line.slice(1) })
    } else if (line.charAt(1) === '[') {
      this.entries.push({ line, type: 'anchor', value: line.slice(2, -2) })
    } else {
      this.entries.push({ line, type: 'attrlist', value: line.slice(1, -1) })
    }
  }
}

class ContainerStack {
  constructor (blockMetadata, tabsDelimiter) {
    this.blockMetadata = blockMetadata
    this.tabsDelimiter = tabsDelimiter
    this.current = this.closestWithUnwrap = undefined
    this.stack = []
    this.inTabs = false
  }

  get verbatim () {
    return !!this.current?.verbatim
  }

  closeBlock (accum) {
    if (this.inTabs) this.closeTabsBlock(accum)
    const currentBlock = this.pop()
    if (!currentBlock.unwrap) return
    const currentBlockMetadata = currentBlock.metadata
    if (currentBlockMetadata.empty) return 'unwrapped'
    let inheritedMetadataLines = currentBlockMetadata.getLines()
    const start = currentBlock.start
    if (currentBlock.multiple === true) {
      accum.splice(start, 0, ...inheritedMetadataLines.concat(currentBlock.delimiter))
      return
    }
    const inheritedTitle = currentBlock.metadata.title
    if (inheritedTitle && accum[start] !== '[tabs]') {
      let titleLineIdx = start
      let l = accum[titleLineIdx]
      if (l && l.charAt() === '[' && l.charAt(l.length - 1) === ']') l = accum[++titleLineIdx]
      if (l && l.charAt() === '.' && l.length > 1 && l.charAt(1) !== ' ' && !l.startsWith('...')) {
        accum[titleLineIdx] = '.' + inheritedTitle + ' - ' + accum[titleLineIdx].slice(1)
        inheritedMetadataLines = currentBlockMetadata.getLines(false, (it) => it.type !== 'title')
        if (!inheritedMetadataLines.length) return 'unwrapped'
      }
    }
    accum.splice(start, 0, ...inheritedMetadataLines)
    return 'unwrapped'
  }

  closeTabsBlock (accum) {
    this.inTabs = false
    if (!accum) return '\n' + this.tabsDelimiter
    const restore = []
    if (accum[accum.length - 1] === '') {
      restore.push(accum.pop())
      while (accum[accum.length - 1] === '') restore.push(accum.pop())
    }
    restore.length ? accum.push(this.tabsDelimiter, ...restore) : accum.push(this.tabsDelimiter)
  }

  closeDanglingBlock (accum) {
    if (this.inTabs) this.closeTabsBlock(accum)
    const { delimiter, metadata, multiple, start, unwrap } = this.current
    if (!unwrap || !metadata.empty || multiple === true) accum.splice(start, 1, '//' + delimiter)
    this.blockMetadata.drain(accum)
    this.pop()
  }

  isNewTabs () {
    return this.inTabs && this.isTab('primary')
  }

  isTab (role) {
    return this.blockMetadata.isTab(role || (this.inTabs ? 'secondary' : 'primary'))
  }

  isTerminus (line) {
    return this.current && line === this.current.delimiter
  }

  openTabsBlock (accum) {
    this.inTabs = true
    accum.push('[tabs]', this.tabsDelimiter)
  }

  pop () {
    this.inTabs &&= 'expecting-tab'
    this.current = this.stack[this.stack.length - 2]
    const previous = this.stack.pop()
    if (previous.tabsDelimiter) this.tabsDelimiter = previous.tabsDelimiter
    if (previous === this.closestWithUnwrap) this.closestWithUnwrap = this.current?.unwrap ? this.current : undefined
    return previous
  }

  push (line, accum, props = {}) {
    const entry = Object.assign({ delimiter: line, metadata: this.blockMetadata.clone(), start: accum.length }, props)
    if (entry.unwrap) {
      this.blockMetadata.clear()
      this.closestWithUnwrap = entry
    } else if (line === this.tabsDelimiter) {
      this.tabsDelimiter = (entry.tabsDelimiter = line) === '====' ? '======' : '===='
    }
    this.stack.push((this.current = entry))
  }

  recordNonTab (accum) {
    if (this.closestWithUnwrap) this.closestWithUnwrap.multiple &&= true
    this.blockMetadata.drain(accum)
  }

  startNewBlock (accum) {
    this.inTabs &&= 'expecting-tab'
    if (this.closestWithUnwrap) this.closestWithUnwrap.multiple ||= 'candidate'
  }

  startTab (accum) {
    if (!this.inTabs) {
      this.openTabsBlock(accum)
    } else {
      this.inTabs = true
      if (accum[accum.length - 1]?.length) accum.push('')
    }
    accum.push(this.blockMetadata.title + '::', '+', ...this.blockMetadata.getLines(true, (it) => it.type !== 'title'))
  }

  shouldUnwrapBlock (mode, lines, idx) {
    if (mode === 'never' || this.blockMetadata.hasStyle()) return false
    if (mode === 'always') return true
    let l = lines[++idx]
    if (l && l.charAt() === '.' && l.length > 1 && l.charAt(1) !== ' ' && !l.startsWith('...')) l = lines[++idx]
    return l && l.charAt() === '[' && l.charAt(l.length - 1) === ']' && ~l.indexOf('role="primary"')
  }
}
