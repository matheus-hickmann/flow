# CLAUDE.md - Full Stack Guide (Spring Boot & Angular)

## 👤 Role & Standards
Expert Full Stack Developer focused on Java (Spring Boot 3) and Angular (17+).
Adheres to SOLID principles, Clean Code, and High-Performance patterns.

---

## 🏗 GLOBAL STANDARDS
- **Indentation**: 2 spaces.
- **Quotes**: Single quotes (`'`) for TS/Angular; Double quotes (`"`) for Java.
- **Naming Conventions**:
    - File Names: `kebab-case` (e.g., `user-service.java`, `account-list.component.ts`).
    - Classes/Interfaces: `PascalCase`.
    - Methods/Variables: `camelCase`.
    - Constants: `ALL_CAPS_WITH_UNDERSCORES`.

---

## ☕ BACKEND: Java & Spring Boot 3
- **Injection**: Use **Constructor Injection** only. Avoid `@Autowired`.
- **Architecture**: Controller -> Service -> Repository -> Entity/Model.
- **Features**: Java 17+ (Records for DTOs, Sealed Classes, Pattern Matching).
- **Persistence**: Spring Data JPA. Use Flyway/Liquibase if requested.
- **Testing**:
    - Frameworks: JUnit 5, Mockito, **AssertJ**.
    - Pattern: `classname_methodName_testCase`.
    - Assertions: Always use AssertJ `assertThat()`.
- **Validation**: Use `jakarta.validation` (`@Valid`).

---

## 🅰 FRONTEND: Angular 17+
- **Reactivity**: Use **Signals** for state management. Avoid manual RxJS subscriptions in components.
- **DI**: Use the `inject()` function instead of constructor parameters.
- **Components**: Use **Standalone Components** exclusively.
- **Templates**: Use `async` pipe and `@defer` for lazy loading.
- **Performance**: Use `NgOptimizedImage` and `trackBy` in loops.
- **Styling**: Tailwind CSS for utility-first design; SASS for custom components.

---

## 🛠 COMMANDS & WORKFLOW

### Backend (Maven)
- Build: `./mvnw clean install`
- Run: `./mvnw spring-boot:run`
- Test: `./mvnw test`

### Frontend (NPM)
- Install: `npm install`
- Start: `npm start`
- Test: `npm test`
- Build: `npm run build`

---

## 🧪 TESTING & QUALITY
- **TDD**: Implement TDD for Services and Logic classes.
- **Exclusions**: Do NOT test DTOs, simple Getters/Setters, or boilerplate.
- **Pattern**: Arrange-Act-Assert.
- **Immutability**: Use `final` in Java and `readonly` in TypeScript.

---

## 🚫 ALWAYS AVOID
- Field Injection in Spring Boot.
- Using `any` in TypeScript (Strict typing required).
- Business logic in Controllers or Angular Templates.
- Direct DOM manipulation in Angular.
- Hardcoded configurations (Use `application.yml` or `environment.ts`).