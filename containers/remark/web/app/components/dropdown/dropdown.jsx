/** @jsx h */
import { Component, h } from 'preact';

import Button from 'components/button';

export default class Dropdown extends Component {
  constructor(props) {
    super(props);

    this.state = {
      isActive: props.isActive || false,
    };

    this.onTitleClick = this.onTitleClick.bind(this);
    this.onOutsideClick = this.onOutsideClick.bind(this);
    this.receiveMessage = this.receiveMessage.bind(this);
  }

  onTitleClick() {
    this.setState({
      isActive: !this.state.isActive,
    });

    if (this.props.onTitleClick) {
      this.props.onTitleClick();
    }
  }

  receiveMessage(e) {
    try {
      const data = typeof e.data === 'string' ? JSON.parse(e.data) : e.data;

      if (data.clickOutside) {
        if (this.state.isActive) {
          this.setState({
            isActive: false,
          });
        }
      }
    } catch (e) {}
  }

  onOutsideClick(e) {
    if (!this.rootNode.contains(e.target)) {
      if (this.state.isActive) {
        this.setState({
          isActive: false,
        });
      }
    }
  }

  componentDidMount() {
    document.addEventListener('click', this.onOutsideClick);

    window.addEventListener('message', this.receiveMessage);
  }

  componentWillUnmount() {
    document.removeEventListener('click', this.onOutsideClick);

    window.removeEventListener('message', this.receiveMessage);
  }

  render(props, { isActive }) {
    const { title, heading, children, mix, mods } = props;

    return (
      <div className={b('dropdown', { mix, mods }, { active: isActive })} ref={r => (this.rootNode = r)}>
        <Button
          aria-haspopup="listbox"
          aria-expanded={isActive && 'true'}
          mix="dropdown__title"
          type="button"
          onClick={this.onTitleClick}
        >
          {title}
        </Button>

        <div className="dropdown__content" tabindex="-1" role="listbox">
          {heading && <div className="dropdown__heading">{heading}</div>}
          <div className="dropdown__items">{children}</div>
        </div>
      </div>
    );
  }
}
