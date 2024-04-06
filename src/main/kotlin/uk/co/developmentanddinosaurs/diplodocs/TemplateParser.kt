package uk.co.developmentanddinosaurs.diplodocs

/**
 * Parses a template string into a list of template nodes.
 *
 * Analyzes the input text and generates template nodes based on its content and structure.
 */
class TemplateParser {
    private val pattern = Regex("\\{\\{\\s*(.*?)\\s*}}")

    /**
     * Parses the given template string and returns a list of template nodes.
     *
     * @param template The template to parse.
     * @return A list of template nodes representing the parsed template.
     */
    fun parse(template: String): List<TemplateNode> {
        val nodes = mutableListOf<TemplateNode>()
        var lastIndex = 0

        pattern.findAll(template).forEach { match ->
            val matchIndex = match.range.first
            if (matchIndex > lastIndex) {
                // Add text node for text before placeholder
                val text = template.substring(lastIndex, matchIndex)
                nodes.add(TextNode(text))
            }
            // Add placeholder node for the matched placeholder
            val placeholder = match.groupValues[1]
            require(isValidPlaceholder(placeholder)) { "Invalid placeholder syntax: '{{ $placeholder }}'" }
            nodes.add(PlaceholderNode(placeholder))
            lastIndex = match.range.last + 1
        }
        // Add text node for text after last placeholder
        if (lastIndex < template.length) {
            val text = template.substring(lastIndex)
            nodes.add(TextNode(text))
        }
        return nodes
    }

    private fun isValidPlaceholder(placeholder: String): Boolean {
        return placeholder.isNotEmpty() && placeholder.matches(Regex("\\w+(\\.\\w+)*"))
    }
}
