plugins {
    kotlin("jvm") version "1.9.23"
    id("com.diffplug.spotless") version "6.25.0"
}

group = "uk.co.developmentanddinosaurs"
version = "1.0-SNAPSHOT"

repositories {
    mavenCentral()
}

dependencies {
    testImplementation(kotlin("test"))
    testImplementation("io.kotest:kotest-runner-junit5-jvm:5.8.1")
    testImplementation("io.kotest:kotest-assertions-core-jvm:5.8.1")
}

spotless {
    kotlin {
        target("src/**/*.kt")
        ktlint()
    }
    kotlinGradle {
        ktlint()
    }
    freshmark {
        target("**/*.md")
    }
    yaml {
        target("**/*.yml")
        prettier()
    }
}

tasks.test {
    useJUnitPlatform()
}

kotlin {
    jvmToolchain(17)
}
