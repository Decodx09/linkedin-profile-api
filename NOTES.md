# What this project is (in plain English)

The goal was simple to describe: you give it a LinkedIn profile link, and it hands
you back all the useful stuff from that profile — name, headline, location, the
"About" text, jobs, education, skills, certifications, languages, photos — as clean,
organized data instead of a messy web page.

Think of it like a waiter: you tell it "table 5, the Bill Gates profile," and it
goes to the kitchen (LinkedIn), grabs the plate, and brings it back neatly.

---

## What we built

A small web service written in **Go** (using a framework called **Fiber**). You run
it, and it gives you a web address you can call, like:

```
/api/profile?url=https://www.linkedin.com/in/some-person/
```

and it replies with the profile details in a tidy format (JSON).

It also has the sensible extras you'd expect:
- A health check (so a hosting service knows it's alive).
- An optional password/key so not just anyone can use your copy.
- A short-term memory (cache) so it doesn't re-ask LinkedIn the same thing over and over.
- Ready-to-go setup files to put it online (Docker, Render).
- A written guide (the main README) and automated tests for the data-organizing part.

**Important rule we followed:** no passwords or secret keys are stored in the code.
They stay in a private settings file that never gets uploaded.

---

## What works today

- The service itself runs perfectly. Every button and error message does the right thing.
- Logging in to LinkedIn with your account cookie **works** — we confirmed it pulled
  back the correct account ("Shivansh", the right headline) during testing.
- The code is clean, tested, and published to GitHub.

## What does NOT work yet (being honest)

We could log in, but we could **not** pull a full profile. Two reasons:

1. **LinkedIn changed the doorway.** The specific "door" our code was knocking on to
   get profile details has been permanently closed by LinkedIn (they replaced it with
   a newer one). So that part of the code needs to be pointed at the new door.

2. **LinkedIn actively fights automation.** LinkedIn is very good at spotting when a
   program — not a real person in a browser — is asking for data. After only a handful
   of automated requests during testing, LinkedIn put a **"verify it's you" security
   check** on the account and started blocking the automated requests (while a normal
   browser still worked fine). This is a well-known challenge with anything that reads
   LinkedIn automatically, and it's the main obstacle, not the code.

---

## What it would take to fully finish

1. **Clear the security check** on the LinkedIn account by logging in normally in a
   browser and completing the "verify it's you" prompt.
2. **Grab one real example** of the new profile data straight from the browser (which
   LinkedIn trusts), so the code can be taught the exact new format.
3. **Point the code at LinkedIn's new doorway** and teach it to read the new format.
4. **Add disguise + patience** so it looks more like a real browser and asks slowly, to
   avoid tripping LinkedIn's alarms. For heavy use, you'd also route it through rotating
   internet addresses (proxies).

None of these are dead-ends — they're the normal, expected steps for this kind of tool.
We stopped here on purpose rather than keep triggering LinkedIn's security checks or
ship something that only looks finished but hasn't actually pulled a real profile.

---

## The short version

The "waiter" is hired, trained, and shows up for work. It can walk into the kitchen
(log in). But LinkedIn just moved the kitchen around and put a guard at the door for
anyone who isn't a regular customer walking in person. Finishing the job means learning
the new kitchen layout and dressing the waiter up to look like a regular — both very
doable, and clearly mapped out above.
