package uk.co.developmentanddinosaurs.diplodocs

/**
 * Parses a template string into a list of template nodes.
 *
 * Analyzes the input text and generates template nodes based on its content and structure.
 */
class TemplateParser {
    /**
     * Parses the given template string and returns a list of template nodes.
     *
     * @param template The template to parse.
     * @return A list of template nodes representing the parsed template.
     */
    fun parse(template: String): List<TemplateNode> {
        return listOf(TextNode(template))
    }
}
