#!/usr/bin/env bash
#
# Renders and copies documentation files into the informed RELEASE_DIR, the script search for
# task templates on a specific glob expression. The templates are rendered using the actual
# task name and documentation is searched for and copied over to the task release directory.
#

shopt -s inherit_errexit
set -eu -o pipefail

readonly RELEASE_DIR="${1:-}"
export IMG=${IMG:-quay.io/openshift-pipeline/pipelines-cache-rhel9:next}

# Print error message and exit non-successfully.
panic() {
    echo "# ERROR: ${*}"
    exit 1
}

# Extracts the filename only, without path or extension.
extract_name() {
    declare filename=$(basename -- "${1}")
    declare extension="${filename##*.}"
    echo "${filename%.*}"
}

# Function to find the respective documentation for a given name, however, for s2i it only consider the
## "task-s2i" part instead of the whole name.
find_doc() {
    declare name="${1}"
    [[ "${name}" == "task-s2i"* ]] &&
        name="task-s2i"
    find docs/ -name "${name}*.md"
}

#
# Main
#

release() {
    # making sure the release directory exists, this script should only create releative
    # directories using it as root
    [[ ! -d "${RELEASE_DIR}" ]] &&
        panic "Release dir is not found '${RELEASE_DIR}'!"

    # Release task templates
#    release_templates "task" "templates/task-*.yaml" "tasks"

    # Release StepAction templates
    release_templates "stepaction" "tekton/cache*.yaml" "stepactions"
}

release_templates() {
    local template_type=$1
    local glob_expression=$2
    local release_subdir=$3

    # releasing all templates using the following glob expression
    for t in $(ls -1 ${glob_expression}); do
        declare name=$(extract_name ${t})
          [[ -z "${name}" ]] &&
              panic "Unable to extract name from '${t}'!"

        echo "Updating Image to $IMG in $t"
        yq -i '.spec.image = env(IMG)' "$t"
        local dir="${RELEASE_DIR}/${release_subdir}/${name}"
          [[ ! -d "${dir}" ]] &&
              mkdir -p "${dir}"

        # rendering the helm template for the specific file, using the resource name for the
        # filename respectively
        echo "# Rendering '${name}' at '${dir}'..."
       cp $t  ${dir}/${name}.yaml ||
            panic "Unable to render '${t}'!"
    done
}
release
