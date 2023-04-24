Q1:
Condition variables, Java monitors, and Ada protected objects are quite similar in what they do (temporarily yield execution so some other task can unblock us).

But in what ways do these mechanisms differ?

Answer:
    Conditions variables makes a thread wait until a certain condition becomes true. The thread that sets the condition to true, or signals the condition, notifies the waiting thread to wake up and check the condition again (thereby "signal"). However, Java monitors and Ada protected objects provides mutual exclusion (mutex) and synchronization.
    
    Java monitors are code blocks that can be accessed only one thread at a time. When a thread enters the monitor, it holds the monitor lock until it exits the monitor. Other threads attempting to access the monitor will be blocked until the lock is released.

    Ada protected objects use the "protected" function to create a block of code that can only be accessed by one thread at a time, similar to Java monitors.


Q2:
Bugs in this kind of low-level synchronization can be hard to spot.

Which solutions are you most confident are correct?
Why, and what does this say about code quality?

Answer:
    All three syncronization mechanisms are safe to use and correct, and choosing the best one depends on the overall design of your system. However, I would be most confident using Java monitor, as it is built into the Java language and provides clear boundaries for where it is implemented.
    Condition variables and Ada protected objects would require better code structure and smart design by the developer.


Q3:
We operated only with two priority levels here, but it makes sense for this "kind" of priority resource to support more priorities.

How would you extend these solutions to N priorities? Is it even possible to do this elegantly?
What (if anything) does that say about code quality?

Answer:
    We could extend the code to support more priorities by using a priority queue supporting more priority levels. This could be done elegantly, but it could also introduce more complexity and and potential for more bugs.
    Of course, extending the code is easier if the code quality is good and the system is well modulized.


Q4:
In D's standard library, getValue for semaphores is not even exposed (probably because it is not portable – Windows semaphores don't have getValue, though you could hack it together with ReleaseSemaphore() and WaitForSingleObject()).

A leading question: Is using getValue ever appropriate?
Explain your intuition: What is it that makes getValue so dubious?

Answer:
    Using getValue on sempahore is dubious because it can lead to race conditions, as the semaphore's state can change while its value is read. Therefore, it is always better to use the standard semaphore library's functions.


Q5:
Which one(s) of these different mechanisms do you prefer, both for this specific task and in general? (This is a matter of taste – there are no "right" answers here)

Answer:
    I am most familiar with C, so I would be most comfortable using condition variables. However, I like the principles of Java Monitors, so I might be interested in learning more about this mechanism in the future.