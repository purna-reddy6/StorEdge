FROM maven:3.9-eclipse-temurin-21 AS builder
WORKDIR /app
COPY pom.xml .
RUN mvn dependency:go-offline -B
COPY src ./src
RUN mvn clean package -DskipTests -B

FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
RUN addgroup -S storedge && adduser -S storedge -G storedge
COPY --from=builder /app/target/*.jar app.jar
USER storedge
EXPOSE 8081
ENTRYPOINT ["java", "-jar", "app.jar"]
