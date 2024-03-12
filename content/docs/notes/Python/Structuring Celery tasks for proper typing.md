# Structuring Celery Beat code to avoid dumb mistakes

...with a solution that isn't very smart 😎

## Problem

Celery uses JSON for serialising information/kwargs that are passed to a task. Normally, using `task.delay()` type syntax, this isn't a problem and you don't need to think about this too much since you can declare the appropriate keyword arguments directly in the function call which will guarantee(ish) that you are supplying the required arguments, and in the correct type.

However, when using `PeriodicTask` to schedule other tasks, the reference to the task to be executed at a certain time is stored as strings, consisting of:

- Name of the task in full, e.g. `core.tasks.my_task`
- `kwargs` to be passed to the task, stored as string, in JSON syntax, e.g. `{"my_argument":"123"}`

In code, declaring a `PeriodicTask` looks something like this:

```python
task = PeriodicTask(
    name="core.tasks.my_task"
    kwargs={
        "my_argument_one" : str(some_value),
        "my_argument_two" : str(another_value),
    },
    ...
)
```

..which means that **you can freely screw up the task name, or its kwargs, and you won't know until your task inevitably fails at runtime, long after your deploymemnt.** As an added bonus, you may also mistype the name of the task.

In my current codebase, tasks were sitting in `tasks.py`, whereas creating this `PeriodicTask` was instead in a different file, triggered by another function.

## Solution (ish)

The best solution I have come up with is to **move the code that creates `PeriodicTasks` as close as possible to the task it is scheduling** to try to get better locality of the behavior. In my case, this meant I created a class called `Tasker` to handle the scheduling, and declared the main task it would call right below it, like so:

```python
# core.tasks.py
class Tasker:
    def __init__(...):
        # init kwargs collect all info I need to do scheduling
        # as I don't want to do a DB lookup

    def schedule_closure(self) -> None:
        ...
        task = PeriodicTask(
            name=f"Run task {self.event_id}",
            task=f"{__name__}.{my_example_task.__name__}",
            kwargs=json.dumps(
                {
                "event_id": str(self.event_id),
                "user_id": str(self.user_id),
                }
            ),
            ...

# and right below, the actual task def 
@app.task()
def my_example_task(event_id: str, user_id: str) -> None:
    ...
```

And elsehwere in the triggering code, I now just have:

```python
tasker = Tasker(
    event_id=self.event.id,
    user_id=self.user.id,
)
tasker.schedule_closure()
```

Better, but still not great. Will need to explore further ways to define a contract that both sides must follow, in a more enforceable way.
