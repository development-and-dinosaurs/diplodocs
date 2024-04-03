package uk.co.developmentanddinosaurs.diplodocs

/**
 * Entry point for template processing operations.
 *
 * Orchestrates the processing of template nodes using a visitor pattern.
 */
class TemplateEngine {
    /**
     * Processes the template nodes using the specified visitor.
     * @param nodes The list of template nodes to process.
     * @param visitor The visitor responsible for processing the nodes.
     * @return The processed template.
     */
    fun processTemplate(
        nodes: List<TemplateNode>,
        visitor: TemplateVisitor,
    ): String {
        val stringBuilder = StringBuilder()
        for (node in nodes) {
            stringBuilder.append(node.accept(visitor))
        }
        return stringBuilder.toString()
    }
}
