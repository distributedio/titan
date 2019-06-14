# Commit Message Style

It is friendly to have a good commit message style for someone who want to learn about the commitment history.
We follow the style of [AngularJS](https://docs.google.com/document/d/1QrDFcIiPjSLDn3EL15IJygNPiHORgU1_OOAqWjiDU5Y/edit#heading=h.greljkmo14y0) and supply more details specific to this project.

## Format of the commit message

```html
<type>(<scope>): <subject>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>
```

Any line of the commit message cannot be longer 100 characters! This allows the message to be easier to read on github as well as in various git tools.

A commit message consists of a header, a body and a footer, separated by a blank line.

## Revert

If the commit reverts a previous commit, its header should begin with `revert: `, followed by the header of the reverted commit. In the body it should say: `This reverts commit <hash>.`, where the hash is the SHA of the commit being reverted.

## Message header

The message header is a single line that contains succinct description of the change containing a **type**, an optional **scope** and a **subject**.

### Allowed `<type>`

This describes the kind of change that this commit is providing.

* feat    (feature)
* fix     (bug fix)
* docs    (documentation)
* style   (formating, missing semi colons, ...)
* refactor
* test    (when adding missing tests)
* chore   (maintain)

### Available `<scope>`

* command
* db
* server
* resp
* context
* conf
* metric
* ...

You can define scope yourself which is not listed here, and you can use * if there isn't a more fitting scope.

### `<subject>` text

This is a very short description of the change.

* use imperative, present tense: “change” not “changed” nor “changes”
* don't capitalize first letter
* no dot (.) at the end

## Message body

* just as in `<subject>` use imperative, present tense: “change” not “changed” nor “changes”
* includes motivation for the change and contrasts with previous behavior

## Message footer

### Breaking changes

All breaking changes have to be mentioned as a breaking change block in the footer, which should start with the word BREAKING CHANGE: with a space or two newlines. The rest of the commit message is then the description of the change, justification and migration notes.

### Referencing issues

Closed bugs should be listed on a separate line in the footer prefixed with "Closes" keyword like this:

```text
Closes #234
```

or in case of multiple issues:

```text
Closes #123, #245, #992
```

## Examples

```text
feat(command): implement the touch command
```

```text
feat(db): remove len from the metadata

It is a heavy performance overhead to update the metakey
each time when modifying an object because of the transaction
conflicts. Having no lengh of an object, it may affect
the performance of commands that need this information.

BREAKING CHANGE:

The layout of the object metadata has been changed. It is not compatible with previous versions.
```
