package uk.co.developmentanddinosaurs.diplodocs

fun main() {
    val templateString = "Hello, my name is {{user.name}}! {% if {{user.name}} == John %} Hi John! {% endif %}"
    val parser = TemplateParser()
    val templateNodes = parser.parse(templateString)

    val engine = TemplateEngine()
    val data = mapOf("user" to mapOf("name" to "John"), "name" to "Beep")
    val processor = TemplateProcessor(data)

    println(engine.processTemplate(templateNodes, processor))
}
