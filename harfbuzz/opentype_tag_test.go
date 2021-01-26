package harfbuzz

test_langs_sorted ()
{
  for (unsigned int i = 1; i < ARRAY_LENGTH (ot_languages); i++)
  {
	int c = ot_languages[i].cmp (&ot_languages[i - 1]);
	if (c > 0)
	{
	  fprintf (stderr, "ot_languages not sorted at index %d: %s %d %s\n",
		   i, ot_languages[i-1].language, c, ot_languages[i].language);
	  abort();
	}
  }
}