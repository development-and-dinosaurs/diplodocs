package uk.co.developmentanddinosaurs.diplodocs

interface TemplateVisitor {
    fun visit(textNode: TextNode): String
}
