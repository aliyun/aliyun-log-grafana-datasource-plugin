import { css, cx } from '@emotion/css';
import React, { Component, HTMLAttributes } from 'react';
import { withTheme2, ReactUtils } from '@grafana/ui'; // Assuming you export GrafanaTheme2 and withTheme2 from @grafana/ui
import { GrafanaTheme2 } from '@grafana/data';

import  FieldValidationMessage  from './FieldValidationMessage';
import  Label  from './Label';

export interface FieldProps extends HTMLAttributes<HTMLDivElement> {
  /** Form input element, i.e Input or Switch */
  children: React.ReactElement;
  /** Label for the field */
  label?: React.ReactNode;
  /** Description of the field */
  description?: React.ReactNode;
  /** Indicates if field is in invalid state */
  invalid?: boolean;
  /** Indicates if field is in loading state */
  loading?: boolean;
  /** Indicates if field is disabled */
  disabled?: boolean;
  /** Indicates if field is required */
  required?: boolean;
  /** Error message to display */
  error?: React.ReactNode;
  /** Indicates horizontal layout of the field */
  horizontal?: boolean;
  /** make validation message overflow horizontally. Prevents pushing out adjacent inline components */
  validationMessageHorizontalOverflow?: boolean;

  className?: string;
  /**
   *  A unique id that associates the label of the Field component with the control with the unique id.
   *  If the `htmlFor` property is missing the `htmlFor` will be inferred from the `id` or `inputId` property of the first child.
   *  https://developer.mozilla.org/en-US/docs/Web/HTML/Element/label#attr-for
   */
  htmlFor?: string;
  theme: GrafanaTheme2; // Add theme to props for withTheme2 HOC
}

class Field extends Component<FieldProps> {
  static defaultProps = {
    validationMessageHorizontalOverflow: false,
  };

  getStyles = () => {
    const { theme } = this.props;
    return {
      field: css`
        display: flex;
        flex-direction: column;
        margin-bottom: ${theme.spacing(2)};
      `,
      fieldHorizontal: css`
        flex-direction: row;
        justify-content: space-between;
        flex-wrap: wrap;
      `,
      fieldValidationWrapper: css`
        margin-top: ${theme.spacing(0.5)};
      `,
      fieldValidationWrapperHorizontal: css`
        flex: 1 1 100%;
      `,
      validationMessageHorizontalOverflow: css`
        width: 0;
        overflow-x: visible;

        & > * {
          white-space: nowrap;
        }
      `,
    };
  };

  deleteUndefinedProps = <T extends Object>(obj: T): Partial<T> => {
    for (const key in obj) {
      if (obj[key] === undefined) {
        delete obj[key];
      }
    }
    return obj;
  };

  render() {
    const {
      label,
      description,
      horizontal,
      invalid,
      loading,
      disabled,
      required,
      error,
      children,
      className,
      validationMessageHorizontalOverflow,
      htmlFor,
      theme,
      ...otherProps
    } = this.props;

    const styles = this.getStyles();
    const inputId = htmlFor ?? ReactUtils?.getChildId?.(children);

    const labelElement =
      typeof label === 'string' ? (
        <Label htmlFor={inputId} description={description}>
          {`${label}${required ? ' *' : ''}`}
        </Label>
      ) : (
        label
      );

    const childProps = this.deleteUndefinedProps({ invalid, disabled, loading });

    return (
      <div className={cx(styles.field, horizontal && styles.fieldHorizontal, className)} {...otherProps}>
        {labelElement}
        <div>
          {React.cloneElement(children, childProps)}
          {invalid && error && !horizontal && (
            <div
              className={cx(styles.fieldValidationWrapper, {
                [styles.validationMessageHorizontalOverflow]: !!validationMessageHorizontalOverflow,
              })}
            >
              <FieldValidationMessage>{error}</FieldValidationMessage>
            </div>
          )}
        </div>

        {invalid && error && horizontal && (
          <div
            className={cx(styles.fieldValidationWrapper, styles.fieldValidationWrapperHorizontal, {
              [styles.validationMessageHorizontalOverflow]: !!validationMessageHorizontalOverflow,
            })}
          >
            <FieldValidationMessage>{error}</FieldValidationMessage>
          </div>
        )}
      </div>
    );
  }
}

export default withTheme2(Field);
