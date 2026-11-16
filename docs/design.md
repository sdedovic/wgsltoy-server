# Design

## Repository method names

Repository methods follow [Spring Data JPA's query method names](https://docs.spring.io/spring-data/jpa/reference/repositories/query-methods-details.html), so a method's name says what it queries.

- `Save(ctx, entity)` inserts or updates.
- `FindByX(ctx, x)` returns one entity, or `ErrNotFound`.
- `FindAllByX(ctx, x)` returns a slice, empty when nothing matches.
- `ExistsByX`, `CountByX`, and `DeleteByX` return a `bool`, a count, and nothing.
- Conditions combine with `And` and `Or`, and take operators such as `In`, `Not`, and `IsNull`: `FindAllByVisibilityIn`.
- `TopN` limits and `OrderByXAsc`/`OrderByXDesc` sorts: `FindTop100ByCreatedByOrderByUpdatedAtDesc`.
- `X` is the entity's Go field name (`CreatedBy`), not the column name.
