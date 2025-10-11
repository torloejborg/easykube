/**
 * Type-hintable wrapper for exposed GO utility functions
 */
class EasykubeUtils  {
    constructor() {
    }

    /**
     * Generates a new UUID
     * @returns {string}
     */
    newUUID() {
        return _utils.UUID()
    }
}
const utils = new EasykubeUtils();


class Git {
    constructor() {
    }

    /**
     * Downloads a particular directory in a git repository
     * @param {string} repository the git repository
     * @param {string} branch which branch to check out
     * @param {string[]} directories which directories in the repository to check out sparsely
     * @param {string} destination directory for checkout
     */
    sparseCheckout(repository, branch, directories,destination) {
        _ek.sparseCheckout(repository,branch,directories,destination)
    }

}

const git = new Git();


// newline at end-of-file is needed
