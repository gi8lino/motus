require "yaml"

path = File.expand_path("../examples/core-exercises.yaml", __dir__)
data = YAML.safe_load(File.read(path))
abort "catalog version must be 2" unless data["version"] == 2
exercises = data.fetch("exercises")
abort "exercises must be an array" unless exercises.is_a?(Array)
exercises.each_with_index do |exercise, index|
  abort "exercise #{index + 1} must be an object" unless exercise.is_a?(Hash)
  abort "exercise #{index + 1} requires a name" unless exercise["name"].is_a?(String) && !exercise["name"].strip.empty?
  unknown = exercise.keys - %w[name hasSides labels previousNames]
  abort "exercise #{index + 1} has unknown fields: #{unknown.join(', ')}" unless unknown.empty?
  labels = exercise.fetch("labels", [])
  abort "exercise #{index + 1} labels must be non-empty strings" unless labels.is_a?(Array) && labels.all? { |label| label.is_a?(String) && label == label.strip.downcase && !label.empty? }
  previous_names = exercise.fetch("previousNames", [])
  abort "exercise #{index + 1} previousNames must be non-empty strings" unless previous_names.is_a?(Array) && previous_names.all? { |name| name.is_a?(String) && !name.strip.empty? }
end

names = exercises.map { |exercise| exercise.fetch("name").strip.downcase }
abort "exercise names must be unique" unless names.uniq.length == names.length
