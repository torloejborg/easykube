'use strict'

const TagSemVerRx = /^v?(\d+)\.(\d+)\.(\d+)(?:\.(\d+))?(?:-(RC|M)(\d+))?$/

const { name: packageName } = require('#package')
/**
 * Removes components with tags that are not the latest patch of a generation.
 */
module.exports.register = function ({ playbook }) {
  const extensionNames = (playbook?.antora?.extensions || []).map((e) => e.require || e)
  const atlasIndex = extensionNames.findIndex((e) => e === '@antora/atlas-extension')
  const indexOfThis = extensionNames.findIndex((e) => e === `${packageName}/latest-version-extension`)
  if (atlasIndex > -1 && indexOfThis > -1 && indexOfThis < atlasIndex) {
    throw new Error('The latest-version-extension must be registered after the atlas-extension')
  }
  const logger = this.getLogger('latest-version-extension')
  this.once('contentAggregated', ({ contentAggregate }) => {
    const componentToGenerationMapping = new Map()
    const tags = contentAggregate
      .slice()
      .filter((a) => a.origins[0].tag)
      .sort((a, b) => compareSemVerAscending(componentRefname(a), componentRefname(b)))
      .reverse()
    for (const componentVersionBucket of tags) {
      const { name: componentName } = componentVersionBucket
      if (!componentToGenerationMapping.has(componentName)) {
        componentToGenerationMapping.set(componentName, new Map())
      }
      const generationToComponenVersionBucket = componentToGenerationMapping.get(componentName)
      const refname = componentRefname(componentVersionBucket)
      const generation = extractGeneration(refname)
      if (generationToComponenVersionBucket.has(generation)) {
        // the size of contentAggregate changes, so we must look up the index every time
        const indexToRemove = contentAggregate.findIndex((c) => c === componentVersionBucket)
        contentAggregate.splice(indexToRemove, 1)
        continue
      }

      generationToComponenVersionBucket.set(generation, componentVersionBucket)
    }
    logger.info(`Using refnames ${contentAggregate.map((c) => c.name + ' ' + componentRefname(c))}`)
  })

  this.once('componentsRegistered', ({ contentCatalog }) => {
    const components = contentCatalog.getComponents()
    for (const component of components) {
      for (const componentVersion of component.versions) {
        const { version, prerelease } = componentVersion
        const generation = extractGeneration(version)
        const versionSegment = prerelease === '-SNAPSHOT' ? generation + prerelease : generation
        componentVersion.versionSegment = versionSegment
      }
    }
  })
}

function extractGeneration (version) {
  const parts = extractVersionParts(version).slice(0, 2)
  return parts.join('.')
}

function componentRefname (c) {
  // FIXME: Can there be more than a single origin?
  return c.origins[0].refname
}

function compareSemVerAscending (lhs, rhs) {
  const lhsParts = extractVersionParts(lhs)
  const rhsParts = extractVersionParts(rhs)
  for (const [index, leftPart] of lhsParts.entries()) {
    const rightPart = rhsParts[index]
    if (leftPart !== rightPart) {
      if (index !== lhsParts.length - 2) {
        return Number(leftPart) - Number(rightPart)
      }
      return leftPart < rightPart ? -1 : 1
    }
  }
  return 0
}

/**
 * Returns an array of the parsed version [int major, int minor, int patch, int hotfix, String type, int typeVersion].
 *
 * If hotfix is undefined, it is explicitly replaced with -1 to ensure proper sorting.
 *
 * The type is either M (milestones), RC (release candidates), or Z (explicitly replaced for releases to ensure proper
 * sorting of releases).
 *
 * The type version is the count of the type. If it is undefined (only releases), then it remains undefined since
 * this comparison is unnecessary.
 *
 * @param version the version in the format of a tag (e.g. v1.0.0-M1 v1.0.0-RC1 v1.0.0 v1.0.1 v1.0.1.1). The v prefix is
 * optional, so v1.0.0 and 1.0.0 are treated the same.
 * @returns an array of the parsed version used for comparison.
 */
function extractVersionParts (version) {
  const match = version.match(TagSemVerRx)
  if (!match) {
    throw new Error(`Cannot parse version = ${version} with regex ${TagSemVerRx}`)
  }
  const result = Array.from(match)
  result.splice(0, 1)
  const hotfixIndex = result.length - 3
  if (result[hotfixIndex] === undefined) {
    result[hotfixIndex] = -1
  }
  const milestoneRcIndex = result.length - 2
  if (result[milestoneRcIndex] === undefined) {
    result[milestoneRcIndex] = 'Z'
  }
  return result
}
