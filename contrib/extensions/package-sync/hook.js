#!/usr/bin/env node
/**
 * Package Sync Extension for sley
 * Synchronizes version to package.json and other JSON files
 */

const fs = require("fs");
const path = require("path");

/**
 * Get nested property value from object using dot notation
 * @param {Object} obj - The object to traverse
 * @param {string} path - Dot-separated path (e.g., "version" or "metadata.version")
 * @returns {*} The value at the path or undefined
 */
function getNestedValue(obj, path) {
  return path.split(".").reduce((current, key) => current?.[key], obj);
}

/**
 * Set nested property value in object using dot notation
 * @param {Object} obj - The object to modify
 * @param {string} path - Dot-separated path (e.g., "version" or "metadata.version")
 * @param {*} value - The value to set
 */
function setNestedValue(obj, path, value) {
  const keys = path.split(".");
  const lastKey = keys.pop();
  const target = keys.reduce((current, key) => {
    if (!(key in current)) {
      current[key] = {};
    }
    return current[key];
  }, obj);
  target[lastKey] = value;
}

/**
 * Update a JSON file with the new version
 * @param {string} filePath - Absolute path to the JSON file
 * @param {string} version - The new version
 * @param {Array<string>} paths - JSON paths to update (default: ["version"])
 * @returns {Object} Result object with success status and message
 */
function updateJsonFile(filePath, version, paths = ["version"]) {
  try {
    // Check if file exists
    if (!fs.existsSync(filePath)) {
      return {
        success: false,
        message: `File not found: ${filePath}`,
      };
    }

    // Read and parse JSON file
    const fileContent = fs.readFileSync(filePath, "utf8");
    let jsonData;

    try {
      jsonData = JSON.parse(fileContent);
    } catch (parseError) {
      return {
        success: false,
        message: `Invalid JSON in ${filePath}: ${parseError.message}`,
      };
    }

    // Update all specified paths
    const updatedPaths = [];
    for (const jsonPath of paths) {
      const oldValue = getNestedValue(jsonData, jsonPath);
      setNestedValue(jsonData, jsonPath, version);
      updatedPaths.push(`${jsonPath}: ${oldValue} -> ${version}`);
    }

    // Write back to file with formatting
    const indentSize = detectIndentation(fileContent);
    const updatedContent = JSON.stringify(jsonData, null, indentSize) + "\n";
    fs.writeFileSync(filePath, updatedContent, "utf8");

    return {
      success: true,
      message: `Updated ${path.basename(filePath)}: ${updatedPaths.join(", ")}`,
    };
  } catch (error) {
    return {
      success: false,
      message: `Error updating ${filePath}: ${error.message}`,
    };
  }
}

/**
 * Detect indentation style of JSON file
 * @param {string} content - File content
 * @returns {number|string} Indentation (number of spaces or '\t')
 */
function detectIndentation(content) {
  // Try to detect indentation from existing content
  const match = content.match(/^(\s+)"[^"]+"/m);
  if (match) {
    const indent = match[1];
    if (indent.includes("\t")) {
      return "\t";
    }
    return indent.length;
  }
  // Default to 2 spaces
  return 2;
}

/**
 * Main entry point
 */
function main() {
  try {
    // Read JSON input from stdin
    let inputData = "";

    process.stdin.on("data", (chunk) => {
      inputData += chunk;
    });

    process.stdin.on("end", () => {
      try {
        const input = JSON.parse(inputData);

        // Extract required fields
        const version = input.version;
        const projectRoot = input.project_root;
        // For monorepo support: use module_dir if provided, otherwise fall back to project_root
        const moduleDir = input.module_dir || projectRoot;
        const moduleName = input.module_name || null;
        const config = input.config || {};

        // Validate required fields
        if (!version) {
          const result = {
            success: false,
            message: "Missing required field: version",
            data: {},
          };
          console.log(JSON.stringify(result));
          process.exit(1);
        }

        if (!projectRoot) {
          const result = {
            success: false,
            message: "Missing required field: project_root",
            data: {},
          };
          console.log(JSON.stringify(result));
          process.exit(1);
        }

        // Get file configurations with defaults
        const fileConfigs = config.files || [
          { path: "package.json", json_paths: ["version"] },
        ];

        // Process each file
        const results = [];
        let hasError = false;

        for (const fileConfig of fileConfigs) {
          // Determine base directory for file resolution:
          // - If path is absolute, use it directly
          // - If 'relative_to' is specified, respect that setting
          // - Otherwise, use moduleDir (which falls back to projectRoot for single-module projects)
          let baseDir = moduleDir;
          const configPath = fileConfig.path || fileConfig;

          if (path.isAbsolute(configPath)) {
            // Absolute path - use as-is
            baseDir = "";
          } else if (fileConfig.relative_to === "project") {
            // Explicitly relative to project root
            baseDir = projectRoot;
          }
          // else: use moduleDir (default, supports monorepo)

          const filePath = baseDir ? path.join(baseDir, configPath) : configPath;
          const jsonPaths = fileConfig.json_paths || ["version"];

          const result = updateJsonFile(filePath, version, jsonPaths);
          results.push(result.message);

          if (!result.success) {
            hasError = true;
          }
        }

        // Return combined result
        const output = {
          success: !hasError,
          message: results.join("; "),
          data: {
            files_processed: fileConfigs.length,
            version: version,
            module_dir: moduleDir !== projectRoot ? moduleDir : undefined,
            module_name: moduleName,
          },
        };

        console.log(JSON.stringify(output));
        process.exit(hasError ? 1 : 0);
      } catch (parseError) {
        const result = {
          success: false,
          message: `Invalid JSON input: ${parseError.message}`,
          data: {},
        };
        console.log(JSON.stringify(result));
        process.exit(1);
      }
    });

    process.stdin.on("error", (error) => {
      const result = {
        success: false,
        message: `Error reading input: ${error.message}`,
        data: {},
      };
      console.log(JSON.stringify(result));
      process.exit(1);
    });
  } catch (error) {
    const result = {
      success: false,
      message: `Unexpected error: ${error.message}`,
      data: {},
    };
    console.log(JSON.stringify(result));
    process.exit(1);
  }
}

// Run main function
main();
