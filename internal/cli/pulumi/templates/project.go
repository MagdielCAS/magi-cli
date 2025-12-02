package templates

import "fmt"

// GetPulumiYamlTemplate returns a default Pulumi.yaml template
func GetPulumiYamlTemplate(projectName, description string) string {
	return fmt.Sprintf(`name: %s
runtime: nodejs
description: %s
`, projectName, description)
}

// GetPackageJsonTemplate returns a default package.json template
func GetPackageJsonTemplate(projectName string) string {
	return fmt.Sprintf(`{
  "name": "%s",
  "devDependencies": {
    "@types/node": "^16.0.0"
  },
  "dependencies": {
    "@pulumi/pulumi": "^3.0.0",
    "@pulumi/aws": "^5.0.0",
    "@pulumi/awsx": "^1.0.0"
  }
}
`, projectName)
}

// GetTsConfigTemplate returns a default tsconfig.json template
func GetTsConfigTemplate() string {
	return `{
  "compilerOptions": {
    "strict": true,
    "outDir": "bin",
    "target": "es2016",
    "module": "commonjs",
    "moduleResolution": "node",
    "sourceMap": true,
    "experimentalDecorators": true,
    "pretty": true,
    "noFallthroughCasesInSwitch": true,
    "noImplicitReturns": true,
    "forceConsistentCasingInFileNames": true
  },
  "files": [
    "index.ts"
  ]
}
`
}
