/** @jsx h */
import { h, Component } from 'preact';

import api from 'common/api';
import { getHandleClickProps } from 'common/accessibility';
import { API_BASE, BASE_URL, COMMENT_NODE_CLASSNAME_PREFIX, BLOCKING_DURATIONS } from 'common/constants';
import { url } from 'common/settings';
import store from 'common/store';

import Input from 'components/input';

import Avatar from 'components/avatar-icon';

export default class Comment extends Component {
  constructor(props) {
    super(props);

    this.state = {
      isReplying: false,
      isEditing: false,
      isUserVerified: false,
      editTimeLeft: null,
    };

    this.votingPromise = Promise.resolve();

    this.updateState(props);

    this.decreaseScore = this.decreaseScore.bind(this);
    this.increaseScore = this.increaseScore.bind(this);
    this.toggleEditing = this.toggleEditing.bind(this);
    this.toggleReplying = this.toggleReplying.bind(this);
    this.toggleCollapse = this.toggleCollapse.bind(this);
    this.toggleUserInfoVisibility = this.toggleUserInfoVisibility.bind(this);
    this.scrollToParent = this.scrollToParent.bind(this);
    this.onEdit = this.onEdit.bind(this);
    this.onReply = this.onReply.bind(this);
    this.onDeleteClick = this.onDeleteClick.bind(this);
    this.onBlockUserClick = this.onBlockUserClick.bind(this);
    this.onUnblockUserClick = this.onUnblockUserClick.bind(this);
  }

  componentWillReceiveProps(nextProps) {
    this.updateState(nextProps);
  }

  updateState(props) {
    const {
      data,
      data: {
        user: { block, id: commentUserId },
        pin,
      },
      mods: { guest } = {},
    } = props;

    const votes = (data && data.votes) || [];
    const score = (data && data.score) || 0;

    if (this.editTimerInterval) {
      clearInterval(this.editTimerInterval);
      this.editTimerInterval = null;
    }

    if (guest) {
      this.setState({
        guest,
        score,
        deleted: data ? data.delete : false,
      });
    } else {
      const userId = store.get('user').id;

      this.setState({
        guest,
        score,
        pinned: !!pin,
        deleted: data ? data.delete : false,
        userBlocked: !!block,
        scoreIncreased: userId in votes && votes[userId],
        scoreDecreased: userId in votes && !votes[userId],
      });

      if (userId === commentUserId) {
        const editDuration = store.get('config') && store.get('config').edit_duration;
        const getEditTimeLeft = () => Math.floor(editDuration - (new Date() - new Date(data.time)) / 1000);

        if (getEditTimeLeft() > 0) {
          this.editTimerInterval = setInterval(() => {
            const editTimeLeft = getEditTimeLeft();

            if (editTimeLeft < 0) {
              clearInterval(this.editTimerInterval);
              this.editTimerInterval = null;
              this.setState({ editTimeLeft: null });
            } else {
              this.setState({ editTimeLeft });
            }
          }, 1000);
        }
      }
    }
  }

  toggleReplying() {
    const { isReplying } = this.state;
    const onPrevInputToggleCb = store.get('onPrevInputToggleCb');

    this.setState({ isEditing: false }, () => this.setState({ isReplying: !isReplying }));

    if (onPrevInputToggleCb) onPrevInputToggleCb();

    if (!isReplying) {
      store.set('onPrevInputToggleCb', () => this.setState({ isReplying: false }));
    } else {
      store.set('onPrevInputToggleCb', null);
    }
  }

  toggleEditing() {
    const { isEditing } = this.state;
    const onPrevInputToggleCb = store.get('onPrevInputToggleCb');

    this.setState({ isReplying: false }, () => this.setState({ isEditing: !isEditing }));

    if (onPrevInputToggleCb) onPrevInputToggleCb();

    if (!isEditing) {
      store.set('onPrevInputToggleCb', () => this.setState({ isEditing: false }));
    } else {
      store.set('onPrevInputToggleCb', null);
    }
  }

  toggleUserInfoVisibility() {
    if (window.parent) {
      const { user } = this.props.data;
      const data = JSON.stringify({ isUserInfoShown: true, user });
      window.parent.postMessage(data, '*');
    }
  }

  togglePin(isPinned) {
    const { id } = this.props.data;
    const promptMessage = `Do you want to ${isPinned ? 'unpin' : 'pin'} this user?`;

    if (confirm(promptMessage)) {
      this.setState({ pinned: !isPinned });

      (isPinned ? api.unpinComment : api.pinComment)({ id, url }).then(() => {
        api.getComment({ id }).then(comment => store.replaceComment(comment));
      });
    }
  }

  toggleVerify(isVerified) {
    const {
      id,
      user: { id: userId },
    } = this.props.data;
    const promptMessage = `Do you want to ${isVerified ? 'unverify' : 'verify'} this user?`;

    if (confirm(promptMessage)) {
      this.setState({ isUserVerified: !isVerified });

      (isVerified ? api.removeVerifyStatus : api.setVerifyStatus)({ id: userId }).then(() => {
        api.getComment({ id }).then(comment => store.replaceComment(comment));
      });
    }
  }

  onBlockUserClick(e) {
    const {
      id,
      user: { id: userId },
    } = this.props.data;

    const ttl = e.target.value;
    const duration = BLOCKING_DURATIONS.find(el => el.value === ttl).label;
    const promptMessage =
      ttl === 'permanently'
        ? 'Do you want to permanently block this user?'
        : `Do you want to block this user (${duration.toLowerCase()})?`;
    if (confirm(promptMessage)) {
      this.setState({ userBlocked: true });

      api
        .blockUser({ id: userId, ttl })
        .then(api.getComment({ id }))
        .then(comment => store.replaceComment(comment));
    }
  }

  onUnblockUserClick() {
    const {
      id,
      user: { id: userId },
    } = this.props.data;

    const promptMessage = `Do you want to unblock this user?`;

    if (confirm(promptMessage)) {
      this.setState({ userBlocked: false });

      api
        .unblockUser({ id: userId })
        .then(api.getComment({ id }))
        .then(comment => store.replaceComment(comment));
    }
  }

  onDeleteClick() {
    const { id } = this.props.data;

    if (confirm('Do you want to delete this comment?')) {
      this.setState({
        deleted: true,
        isEditing: false,
        isReplying: false,
      });

      api.removeComment({ id }).then(() => {
        api.getComment({ id }).then(comment => store.replaceComment(comment));
      });
    }
  }

  increaseScore() {
    const { score, scoreIncreased, scoreDecreased } = this.state;
    const { id } = this.props.data;

    if (scoreIncreased) return;

    this.setState({
      scoreIncreased: !scoreDecreased,
      scoreDecreased: false,
      score: score + 1,
    });

    this.votingPromise = this.votingPromise.then(() => {
      return api.putCommentVote({ id, url, value: 1 }).then(() => {
        api.getComment({ id }).then(comment => store.replaceComment(comment));
      });
    });
  }

  decreaseScore() {
    const { score, scoreIncreased, scoreDecreased } = this.state;
    const { id } = this.props.data;

    if (scoreDecreased) return;

    this.setState({
      scoreDecreased: !scoreIncreased,
      scoreIncreased: false,
      score: score - 1,
    });

    this.votingPromise = this.votingPromise.then(() => {
      return api.putCommentVote({ id, url, value: -1 }).then(() => {
        api.getComment({ id }).then(comment => store.replaceComment(comment));
      });
    });
  }

  onReply(...rest) {
    this.props.onReply(...rest);

    this.setState({
      isReplying: false,
    });
  }

  onEdit(...rest) {
    this.props.onEdit(...rest);

    this.setState({
      isEditing: false,
    });
  }

  scrollToParent(e) {
    const {
      data: { pid },
    } = this.props;

    e.preventDefault();

    const parentCommentNode = document.getElementById(`${COMMENT_NODE_CLASSNAME_PREFIX}${pid}`);

    if (parentCommentNode) {
      parentCommentNode.scrollIntoView();
    }
  }

  toggleCollapse() {
    this.setState({
      isEditing: false,
      isReplying: false,
    });

    if (this.props.onCollapseToggle) {
      this.props.onCollapseToggle();
    }
  }

  render(
    props,
    {
      guest,
      userBlocked,
      pinned,
      score,
      scoreIncreased,
      scoreDecreased,
      deleted,
      isReplying,
      isEditing,
      isUserVerified,
      editTimeLeft,
    }
  ) {
    const { data, mods = {}, isCommentsDisabled } = props;
    const isAdmin = !guest && store.get('user').admin;
    const isGuest = guest || !Object.keys(store.get('user')).length;
    const isCurrentUser = (data.user && data.user.id) === (store.get('user') && store.get('user').id);
    const config = store.get('config') || {};
    const lowCommentScore = config.low_score;

    const o = {
      ...data,
      text: data.text.length
        ? mods.view === 'preview'
          ? getTextSnippet(data.text)
          : data.text
        : userBlocked
          ? 'This user was blocked'
          : deleted
            ? 'This comment was deleted'
            : data.text,
      time: formatTime(new Date(data.time)),
      orig: isEditing
        ? data.orig &&
          data.orig.replace(/&[#A-Za-z0-9]+;/gi, entity => {
            const span = document.createElement('span');
            span.innerHTML = entity;
            return span.innerText;
          })
        : data.orig,
      score: {
        value: Math.abs(score),
        sign: score > 0 ? '+' : score < 0 ? '−' : null,
        view: score > 0 ? 'positive' : score < 0 ? 'negative' : null,
      },
      user: {
        ...data.user,
        picture: data.user.picture.indexOf(API_BASE) === 0 ? `${BASE_URL}${data.user.picture}` : data.user.picture,
        verified: data.user.verified || isUserVerified,
      },
    };

    const defaultMods = {
      pinned,
      // TODO: we also have critical_score, so we need to collapse comments with it in future
      useless: userBlocked || deleted || (score <= lowCommentScore && !mods.pinned && !mods.disabled),
      // TODO: add default view mod or don't?
      view: o.user.admin ? 'admin' : null,
      replying: isReplying,
      editing: isEditing,
    };

    if (mods.view === 'preview') {
      return (
        <article className={b('comment', props, defaultMods)}>
          <div className="comment__body">
            <div className="comment__info">
              <a href={`${o.locator.url}#${COMMENT_NODE_CLASSNAME_PREFIX}${o.id}`} className="comment__username">
                {o.user.name}
              </a>
            </div>{' '}
            <div className={b('comment__text', { mix: 'raw-content' })} dangerouslySetInnerHTML={{ __html: o.text }} />
          </div>
        </article>
      );
    }

    return (
      <article
        className={b('comment', props, defaultMods)}
        id={mods.disabled ? null : `${COMMENT_NODE_CLASSNAME_PREFIX}${o.id}`}
      >
        <div className="comment__body">
          <div className="comment__info">
            {mods.view !== 'user' && <Avatar picture={o.user.picture} />}

            {mods.view !== 'user' && (
              <span
                {...getHandleClickProps(this.toggleUserInfoVisibility)}
                className="comment__username"
                title={o.user.id}
              >
                {o.user.name}
              </span>
            )}

            {isAdmin &&
              mods.view !== 'user' && (
                <span
                  {...getHandleClickProps(() => this.toggleVerify(o.user.verified))}
                  aria-label="Toggle verification"
                  title={o.user.verified ? 'Verified user' : 'Unverified user'}
                  className={b('comment__verification', {}, { active: o.user.verified, clickable: true })}
                />
              )}

            {!isAdmin &&
              !!o.user.verified &&
              mods.view !== 'user' && (
                <span title="Verified user" className={b('comment__verification', {}, { active: true })} />
              )}

            <a href={`${o.locator.url}#${COMMENT_NODE_CLASSNAME_PREFIX}${o.id}`} className="comment__time">
              {o.time}
            </a>

            {mods.level > 0 &&
              mods.view !== 'user' && (
                <a
                  className="comment__link-to-parent"
                  href={`${o.locator.url}#${COMMENT_NODE_CLASSNAME_PREFIX}${o.pid}`}
                  aria-label="Go to parent comment"
                  title="Go to parent comment"
                  onClick={this.scrollToParent}
                >
                  {' '}
                </a>
              )}

            {isAdmin && userBlocked && mods.view !== 'user' && <span className="comment__status">Blocked</span>}

            {isAdmin && !userBlocked && deleted && <span className="comment__status">Deleted</span>}

            {!mods.disabled &&
              mods.view !== 'user' && (
                <span
                  {...getHandleClickProps(this.toggleCollapse)}
                  className={b('comment__action', {}, { type: 'collapse', selected: mods.collapsed })}
                >
                  {mods.collapsed ? '+' : '−'}
                </span>
              )}

            <span className={b('comment__score', {}, { view: o.score.view })}>
              <span
                className={b(
                  'comment__vote',
                  {},
                  { type: 'up', selected: scoreIncreased, disabled: isGuest || isCurrentUser }
                )}
                aria-disabled={isGuest || isCurrentUser}
                {...getHandleClickProps(isGuest || isCurrentUser ? null : this.increaseScore)}
                title={
                  isGuest
                    ? 'Only authorized users are allowed to vote'
                    : isCurrentUser
                      ? "You can't vote for your own comment"
                      : null
                }
              >
                Vote up
              </span>

              <span className="comment__score-value">
                {o.score.sign}
                {o.score.value}
              </span>

              <span
                {...getHandleClickProps(isGuest || isCurrentUser ? null : this.decreaseScore)}
                className={b(
                  'comment__vote',
                  {},
                  { type: 'down', selected: scoreDecreased, disabled: isGuest || isCurrentUser }
                )}
                aria-disabled={isGuest || isCurrentUser ? 'true' : 'false'}
                title={
                  isGuest
                    ? 'Only authorized users are allowed to vote'
                    : isCurrentUser
                      ? "You can't vote for your own comment"
                      : null
                }
              >
                Vote down
              </span>
            </span>
          </div>

          <div className={b('comment__text', { mix: 'raw-content' })} dangerouslySetInnerHTML={{ __html: o.text }} />

          <div className="comment__actions">
            {!deleted &&
              !isCommentsDisabled &&
              !mods.disabled &&
              !isGuest &&
              mods.view !== 'user' && (
                <span {...getHandleClickProps(this.toggleReplying)} className="comment__action">
                  {isReplying ? 'Cancel' : 'Reply'}
                </span>
              )}

            {!deleted &&
              !mods.disabled &&
              !!o.orig &&
              isCurrentUser &&
              (!!editTimeLeft || isEditing) &&
              mods.view !== 'user' && (
                <span
                  {...getHandleClickProps(this.toggleEditing)}
                  className="comment__action comment__action_type_edit"
                >
                  {isEditing ? 'Cancel' : 'Edit'}
                  {editTimeLeft && ` (${editTimeLeft})`}
                </span>
              )}

            {!deleted &&
              isAdmin && (
                <span className="comment__controls">
                  <span {...getHandleClickProps(() => this.togglePin(pinned))} className="comment__control">
                    {pinned ? 'Unpin' : 'Pin'}
                  </span>

                  {userBlocked && (
                    <span {...getHandleClickProps(() => this.onUnblockUserClick())} className="comment__control">
                      Unblock
                    </span>
                  )}

                  {!userBlocked && (
                    <span className="comment__control comment__control_select-label">
                      Block
                      {/* eslint-disable jsx-a11y/no-onchange */}
                      <select className="comment__control_select" onChange={this.onBlockUserClick}>
                        <option disabled selected value>
                          {' '}
                          Blocking period{' '}
                        </option>
                        {BLOCKING_DURATIONS.map(block => <option value={block.value}>{block.label}</option>)}
                      </select>
                    </span>
                  )}

                  {!deleted && (
                    <span {...getHandleClickProps(this.onDeleteClick)} className="comment__control">
                      Delete
                    </span>
                  )}
                </span>
              )}
          </div>
        </div>

        {isReplying &&
          mods.view !== 'user' && (
            <Input mix="comment__input" onSubmit={this.onReply} onCancel={this.toggleReplying} pid={o.id} />
          )}

        {isEditing &&
          mods.view !== 'user' && (
            <Input
              mix="comment__input"
              mods={{ mode: 'edit' }}
              onSubmit={this.onEdit}
              onCancel={this.toggleEditing}
              id={o.id}
              value={o.orig}
              errorMessage={!editTimeLeft && 'Editing time has expired.'}
            />
          )}
      </article>
    );
  }
}

function getTextSnippet(html) {
  const LENGTH = 100;
  const tmp = document.createElement('div');
  tmp.innerHTML = html.replace('</p><p>', ' ');

  const result = tmp.innerText || '';
  const snippet = result.substr(0, LENGTH);

  return snippet.length === LENGTH && result.length !== LENGTH ? `${snippet}...` : snippet;
}

function formatTime(time) {
  // 'ru-RU' adds a dot as a separator
  const date = time.toLocaleDateString(['ru-RU'], { day: '2-digit', month: '2-digit', year: '2-digit' });

  // do it manually because Intl API doesn't add leading zeros to hours; idk why
  const hours = `0${time.getHours()}`.slice(-2);
  const mins = `0${time.getMinutes()}`.slice(-2);

  return `${date} at ${hours}:${mins}`;
}
