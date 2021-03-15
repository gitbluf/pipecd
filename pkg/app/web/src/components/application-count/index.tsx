import React, { FC, memo, useState } from "react";
import {
  Box,
  Card,
  CardActionArea,
  makeStyles,
  Popover,
  Typography,
} from "@material-ui/core";
import { ApplicationKind } from "../../modules/applications";
import { APPLICATION_KIND_TEXT } from "../../constants/application-kind";

const useStyles = makeStyles((theme) => ({
  root: {
    minWidth: 200,
    display: "inline-block",
  },
  actionArea: {
    padding: theme.spacing(2),
  },
  textSpace: {
    marginLeft: theme.spacing(1),
  },
  popover: {
    pointerEvents: "none",
  },
  popoverPaper: {
    padding: theme.spacing(1),
  },
}));

export interface ApplicationCountProps {
  totalCount: number;
  disabledCount: number;
  kind: ApplicationKind;
  onClick: () => void;
}

export const ApplicationCount: FC<ApplicationCountProps> = memo(
  function ApplicationCount({ totalCount, disabledCount, kind, onClick }) {
    const classes = useStyles();

    const [anchorEl, setAnchorEl] = useState<HTMLButtonElement | null>(null);
    const open = Boolean(anchorEl);

    const handlePopoverClose = (): void => {
      setAnchorEl(null);
    };

    return (
      <Card raised className={classes.root}>
        <CardActionArea
          className={classes.actionArea}
          onClick={onClick}
          onMouseEnter={(event) => {
            setAnchorEl(event.currentTarget);
          }}
          onMouseLeave={handlePopoverClose}
        >
          <Typography variant="h6" component="div" color="textSecondary">
            {APPLICATION_KIND_TEXT[kind]}
          </Typography>
          <Box display="flex" justifyContent="center" alignItems="baseline">
            <Typography variant="h4" component="span">
              {totalCount}
            </Typography>
            {disabledCount > 0 ? (
              <Typography
                variant="h6"
                color="textSecondary"
                component="span"
                className={classes.textSpace}
              >
                {`/${disabledCount}`}
              </Typography>
            ) : null}
            <Typography
              variant="h6"
              component="span"
              className={classes.textSpace}
            >
              apps
            </Typography>
          </Box>
        </CardActionArea>

        <Popover
          id="mouse-over-popover"
          className={classes.popover}
          classes={{
            paper: classes.popoverPaper,
          }}
          open={open}
          anchorEl={anchorEl}
          anchorOrigin={{
            vertical: "bottom",
            horizontal: "center",
          }}
          transformOrigin={{
            vertical: "top",
            horizontal: "left",
          }}
          onClose={handlePopoverClose}
          disableRestoreFocus
        >
          <div>
            <b>{totalCount}</b>
            {" total applications"}
          </div>
          <div>
            <b>{disabledCount}</b>
            {" disabled applications"}
          </div>
        </Popover>
      </Card>
    );
  }
);