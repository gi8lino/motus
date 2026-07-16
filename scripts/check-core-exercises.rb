require "yaml"

path = File.expand_path("../examples/core-exercises.yaml", __dir__)
data = YAML.safe_load(File.read(path))
exercises = data.fetch("exercises")
abort "exercises must be an array" unless exercises.is_a?(Array)
exercises.each_with_index do |exercise, index|
  abort "exercise #{index + 1} must be an object" unless exercise.is_a?(Hash)
  abort "exercise #{index + 1} requires a name" unless exercise["name"].is_a?(String) && !exercise["name"].strip.empty?
  unknown = exercise.keys - %w[name hasSides]
  abort "exercise #{index + 1} has unknown fields: #{unknown.join(', ')}" unless unknown.empty?
end
