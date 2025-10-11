'use strict'

module.exports = (
  id,
  {
    data: {
      root: { page },
    },
  }
) => {
  const raw = page.attributes['asciinema-options-' + id]
  if (raw) {
    return raw
  } else {
    return '{}'
  }
}
